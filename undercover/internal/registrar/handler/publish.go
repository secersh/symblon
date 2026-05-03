package handler

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"undercover/internal/registrar/store"
	"undercover/internal/registrar/upload"
	"undercover/pkg/agentpack"
	"undercover/pkg/auth"
)

const maxUploadSize = 20 << 20 // 20 MB

// PublishHandler handles agent package publishing.
type PublishHandler struct {
	agents   store.AgentStore
	uploader upload.Uploader
}

// NewPublishHandler returns a PublishHandler wired to the given store and uploader.
func NewPublishHandler(agents store.AgentStore, uploader upload.Uploader) *PublishHandler {
	return &PublishHandler{agents: agents, uploader: uploader}
}

// Publish handles POST /api/v1/agents.
//
// Expects a multipart/form-data request with a single file field "package"
// containing a zip archive of the agent package directory. The zip must have
// the agent directory as its root entry (e.g. bug-squasher/agent.yaml).
//
// Example:
//
//	zip -r bug-squasher.zip bug-squasher/
//	curl -X POST http://localhost:8082/api/v1/agents \
//	     -H "X-Publisher-ID: secersh" \
//	     -F "package=@bug-squasher.zip"
func (h *PublishHandler) Publish(c *gin.Context) {
	publisher := auth.UserID(c)
	if publisher == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
		return
	}
	publisherName := auth.PublisherName(c)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, _, err := c.Request.FormFile("package")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "package file is required"})
		return
	}
	defer file.Close()

	tmpDir, err := extractZip(file)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": fmt.Sprintf("extract zip: %s", err)})
		return
	}
	defer os.RemoveAll(tmpDir)

	// The zip root may be the package dir itself (e.g. bug-squasher/).
	// Find the directory that contains agent.yaml.
	packageDir, err := findPackageRoot(tmpDir)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	pkg, err := agentpack.Load(packageDir)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	agent, symbols, err := h.publishPackage(c.Request.Context(), publisher, publisherName, pkg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"agent":   agent,
		"symbols": symbols,
		"ref":     fmt.Sprintf("%s/%s/%s", agent.Publisher, agent.Handle, agent.Version),
	})
}

// extractZip writes the contents of the zip to a fresh temp directory and
// returns its path. The caller is responsible for removing it.
func extractZip(r io.ReaderAt) (string, error) {
	// Read into memory so we have a ReaderAt with a known size.
	buf, err := io.ReadAll(io.LimitReader(readerAtToReader(r), maxUploadSize))
	if err != nil {
		return "", err
	}

	zr, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return "", err
	}

	tmpDir, err := os.MkdirTemp("", "agentpack-*")
	if err != nil {
		return "", err
	}

	for _, f := range zr.File {
		if err := extractZipEntry(f, tmpDir); err != nil {
			os.RemoveAll(tmpDir)
			return "", err
		}
	}

	return tmpDir, nil
}

func extractZipEntry(f *zip.File, dest string) error {
	// Prevent zip-slip path traversal.
	target := filepath.Join(dest, filepath.Clean("/"+f.Name))
	if !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
		return fmt.Errorf("invalid zip entry path: %s", f.Name)
	}

	if f.FileInfo().IsDir() {
		return os.MkdirAll(target, 0o755)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	out, err := os.Create(target)
	if err != nil {
		return err
	}
	defer out.Close()

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.Copy(out, rc)
	return err
}

// findPackageRoot returns the directory containing agent.yaml within tmpDir.
// Handles both flat zips (agent.yaml at root) and wrapped zips (dir/agent.yaml).
func findPackageRoot(tmpDir string) (string, error) {
	if _, err := os.Stat(filepath.Join(tmpDir, "agent.yaml")); err == nil {
		return tmpDir, nil
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		candidate := filepath.Join(tmpDir, e.Name())
		if _, err := os.Stat(filepath.Join(candidate, "agent.yaml")); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("agent.yaml not found in zip — expected bug-squasher/agent.yaml or agent.yaml at root")
}

// readerAtToReader wraps an io.ReaderAt as an io.Reader starting at offset 0.
type readerAtToReaderImpl struct {
	r   io.ReaderAt
	off int64
}

func readerAtToReader(r io.ReaderAt) io.Reader {
	return &readerAtToReaderImpl{r: r}
}

func (r *readerAtToReaderImpl) Read(p []byte) (int, error) {
	n, err := r.r.ReadAt(p, r.off)
	r.off += int64(n)
	return n, err
}

func (h *PublishHandler) publishPackage(ctx context.Context, publisher, publisherName string, pkg *agentpack.Package) (*store.Agent, []store.Symbol, error) {
	m := pkg.Manifest

	agent := &store.Agent{
		Publisher:     publisher,
		PublisherName: publisherName,
		Handle:        m.Handle,
		Version:     m.Version,
		Name:        m.Name,
		Description: m.Description,
		Visibility:  string(m.Visibility),
	}

	if m.Pricing != nil {
		agent.PricingModel = string(m.Pricing.Model)
		if m.Pricing.PriceUSD > 0 {
			agent.PriceUSD = &m.Pricing.PriceUSD
		}
	} else {
		agent.PricingModel = string(agentpack.PricingFree)
	}

	symbols := make([]store.Symbol, 0, len(m.Symbols))
	for _, s := range m.Symbols {
		sqlPath, err := h.uploadSQL(ctx, publisher, m.Handle, m.Version, s.ID, pkg.Rules[s.ID])
		if err != nil {
			return nil, nil, fmt.Errorf("upload rule for %q: %w", s.ID, err)
		}

		imageURL := h.uploader.(*upload.S3Uploader).PublicURLForKey(
			fmt.Sprintf("%s/%s/%s/themes/default/%s.svg", publisher, m.Handle, m.Version, s.ID),
		)

		sym := store.Symbol{
			SymbolID:    s.ID,
			Name:        s.Name,
			Description: s.Description,
			Type:        string(s.Type),
			WindowHours: s.WindowHours,
			SQLPath:     sqlPath,
			ImageURL:    imageURL,
		}
		symbols = append(symbols, sym)
	}

	if err := h.uploadAssets(ctx, publisher, m.Handle, m.Version, pkg); err != nil {
		return nil, nil, err
	}

	if err := h.agents.SaveAgent(ctx, agent, symbols); err != nil {
		return nil, nil, fmt.Errorf("save agent: %w", err)
	}

	return agent, symbols, nil
}

func (h *PublishHandler) uploadSQL(ctx context.Context, publisher, handle, version, symbolID, sql string) (string, error) {
	key := fmt.Sprintf("%s/%s/%s/rules/%s.sql", publisher, handle, version, symbolID)
	return h.uploader.Upload(ctx, key, bytes.NewBufferString(sql))
}

func (h *PublishHandler) uploadAssets(ctx context.Context, publisher, handle, version string, pkg *agentpack.Package) error {
	for themeName := range pkg.Themes {
		themeDir := filepath.Join(pkg.Dir, "themes", themeName)
		entries, err := os.ReadDir(themeDir)
		if err != nil {
			return fmt.Errorf("read theme dir %s: %w", themeName, err)
		}
		for _, entry := range entries {
			if entry.IsDir() || filepath.Ext(entry.Name()) != ".svg" {
				continue
			}
			f, err := os.Open(filepath.Join(themeDir, entry.Name()))
			if err != nil {
				return fmt.Errorf("open asset %s: %w", entry.Name(), err)
			}
			defer f.Close()

			key := fmt.Sprintf("%s/%s/%s/themes/%s/%s", publisher, handle, version, themeName, entry.Name())
			if _, err := h.uploader.UploadPublic(ctx, key, f); err != nil {
				return fmt.Errorf("upload asset %s: %w", key, err)
			}
		}
	}
	return nil
}
