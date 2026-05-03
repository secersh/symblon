import * as fs from 'fs';
import * as path from 'path';
import * as archiver from 'archiver';
import { loadToken } from './auth';

const DEFAULT_REGISTRAR = 'https://api.symblon.cc';

function findAgentDir(p: string): string {
  if (fs.existsSync(path.join(p, 'agent.yaml'))) return p;
  for (const entry of fs.readdirSync(p, { withFileTypes: true })) {
    if (entry.isDirectory()) {
      const candidate = path.join(p, entry.name);
      if (fs.existsSync(path.join(candidate, 'agent.yaml'))) return candidate;
    }
  }
  throw new Error(`No agent.yaml found in ${p}`);
}

function zipDir(dir: string): Promise<Buffer> {
  return new Promise((resolve, reject) => {
    const chunks: Buffer[] = [];
    const archive = archiver.default('zip');
    archive.on('data', (chunk: Buffer) => chunks.push(chunk));
    archive.on('end', () => resolve(Buffer.concat(chunks)));
    archive.on('error', reject);
    // Include the agent dir with its folder name as the zip root
    archive.directory(dir, path.basename(dir));
    archive.finalize();
  });
}

export async function publish(inputPath: string): Promise<void> {
  const token = loadToken();
  const registrar = process.env.REGISTRAR_URL ?? DEFAULT_REGISTRAR;

  const agentDir = findAgentDir(path.resolve(inputPath));
  console.log(`Publishing ${path.basename(agentDir)}...`);

  const zip = await zipDir(agentDir);

  const form = new FormData();
  const ab = zip.buffer.slice(zip.byteOffset, zip.byteOffset + zip.byteLength) as ArrayBuffer;
  form.append('package', new Blob([ab], { type: 'application/zip' }), `${path.basename(agentDir)}.zip`);

  const res = await fetch(`${registrar}/registrar/v1/agents`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${token.access_token}` },
    body: form,
  });

  const body = await res.json().catch(() => ({})) as Record<string, unknown>;

  if (res.status === 201) {
    console.log(`Published: ${body.ref}`);
  } else if (res.status === 409) {
    console.log('Already exists — skipping (bump version to republish).');
  } else {
    console.error(`Error (HTTP ${res.status}):`, body);
    process.exit(1);
  }
}
