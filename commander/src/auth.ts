import * as crypto from 'crypto';
import * as fs from 'fs';
import * as http from 'http';
import * as net from 'net';
import * as os from 'os';
import * as path from 'path';

const SUPABASE_URL = 'https://kdddsotxwuuhklojgxrn.supabase.co';

interface StoredToken {
  access_token: string;
  refresh_token: string;
  expires_at: number; // unix ms
}

function tokenPath(): string {
  return path.join(os.homedir(), '.config', 'symblon', 'token.json');
}

export function saveToken(t: StoredToken): void {
  const p = tokenPath();
  fs.mkdirSync(path.dirname(p), { recursive: true, mode: 0o700 });
  fs.writeFileSync(p, JSON.stringify(t, null, 2), { mode: 0o600 });
}

export function loadToken(): StoredToken {
  const p = tokenPath();
  if (!fs.existsSync(p)) {
    console.error('Not logged in. Run: symblon login');
    process.exit(1);
  }
  return JSON.parse(fs.readFileSync(p, 'utf8'));
}

export function logout(): void {
  const p = tokenPath();
  if (fs.existsSync(p)) fs.rmSync(p);
  console.log('Logged out.');
}

function generateVerifier(): string {
  return crypto.randomBytes(32).toString('base64url');
}

function generateChallenge(verifier: string): string {
  return crypto.createHash('sha256').update(verifier).digest('base64url');
}

function freePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const srv = net.createServer();
    srv.listen(0, '127.0.0.1', () => {
      const port = (srv.address() as net.AddressInfo).port;
      srv.close(() => resolve(port));
    });
    srv.on('error', reject);
  });
}

async function exchangeCode(code: string, verifier: string): Promise<StoredToken> {
  const res = await fetch(`${SUPABASE_URL}/auth/v1/token?grant_type=pkce`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ auth_code: code, code_verifier: verifier }),
  });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(`Token exchange failed (${res.status}): ${text}`);
  }
  const data = await res.json() as {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  };
  return {
    access_token: data.access_token,
    refresh_token: data.refresh_token,
    expires_at: Date.now() + data.expires_in * 1000,
  };
}

export async function login(): Promise<void> {
  const verifier = generateVerifier();
  const challenge = generateChallenge(verifier);
  const port = await freePort();
  const redirectURI = `http://127.0.0.1:${port}/callback`;

  const authURL =
    `${SUPABASE_URL}/auth/v1/authorize` +
    `?provider=github` +
    `&redirect_to=${encodeURIComponent(redirectURI)}` +
    `&code_challenge=${challenge}` +
    `&code_challenge_method=s256`;

  const code = await new Promise<string>((resolve, reject) => {
    const timeout = setTimeout(() => {
      srv.close();
      reject(new Error('Timed out waiting for login (2 min)'));
    }, 120_000);

    const srv = http.createServer((req, res) => {
      const url = new URL(req.url!, `http://127.0.0.1:${port}`);
      const code = url.searchParams.get('code');
      if (!code) {
        res.end('Missing code — please try again.');
        return;
      }
      res.writeHead(200, { 'Content-Type': 'text/html' });
      res.end('<html><body><p>Authenticated! You can close this tab.</p></body></html>');
      clearTimeout(timeout);
      srv.close();
      resolve(code);
    });

    srv.listen(port, '127.0.0.1', async () => {
      console.log('Opening browser for GitHub login...');
      const { default: open } = await import('open');
      open(authURL);
      console.log(`\nIf the browser didn't open, visit:\n${authURL}\n`);
    });

    srv.on('error', reject);
  });

  const token = await exchangeCode(code, verifier);
  saveToken(token);
  console.log('Logged in successfully.');
}
