"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.saveToken = saveToken;
exports.loadToken = loadToken;
exports.logout = logout;
exports.login = login;
const crypto = __importStar(require("crypto"));
const fs = __importStar(require("fs"));
const http = __importStar(require("http"));
const net = __importStar(require("net"));
const os = __importStar(require("os"));
const path = __importStar(require("path"));
const SUPABASE_URL = 'https://kdddsotxwuuhklojgxrn.supabase.co';
function tokenPath() {
    return path.join(os.homedir(), '.config', 'symblon', 'token.json');
}
function saveToken(t) {
    const p = tokenPath();
    fs.mkdirSync(path.dirname(p), { recursive: true, mode: 0o700 });
    fs.writeFileSync(p, JSON.stringify(t, null, 2), { mode: 0o600 });
}
function loadToken() {
    const p = tokenPath();
    if (!fs.existsSync(p)) {
        console.error('Not logged in. Run: symblon login');
        process.exit(1);
    }
    return JSON.parse(fs.readFileSync(p, 'utf8'));
}
function logout() {
    const p = tokenPath();
    if (fs.existsSync(p))
        fs.rmSync(p);
    console.log('Logged out.');
}
function generateVerifier() {
    return crypto.randomBytes(32).toString('base64url');
}
function generateChallenge(verifier) {
    return crypto.createHash('sha256').update(verifier).digest('base64url');
}
function freePort() {
    return new Promise((resolve, reject) => {
        const srv = net.createServer();
        srv.listen(0, '127.0.0.1', () => {
            const port = srv.address().port;
            srv.close(() => resolve(port));
        });
        srv.on('error', reject);
    });
}
async function exchangeCode(code, verifier) {
    const res = await fetch(`${SUPABASE_URL}/auth/v1/token?grant_type=pkce`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ auth_code: code, code_verifier: verifier }),
    });
    if (!res.ok) {
        const text = await res.text();
        throw new Error(`Token exchange failed (${res.status}): ${text}`);
    }
    const data = await res.json();
    return {
        access_token: data.access_token,
        refresh_token: data.refresh_token,
        expires_at: Date.now() + data.expires_in * 1000,
    };
}
async function login() {
    const verifier = generateVerifier();
    const challenge = generateChallenge(verifier);
    const port = await freePort();
    const redirectURI = `http://127.0.0.1:${port}/callback`;
    const authURL = `${SUPABASE_URL}/auth/v1/authorize` +
        `?provider=github` +
        `&redirect_to=${encodeURIComponent(redirectURI)}` +
        `&code_challenge=${challenge}` +
        `&code_challenge_method=s256`;
    const code = await new Promise((resolve, reject) => {
        const timeout = setTimeout(() => {
            srv.close();
            reject(new Error('Timed out waiting for login (2 min)'));
        }, 120000);
        const srv = http.createServer((req, res) => {
            const url = new URL(req.url, `http://127.0.0.1:${port}`);
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
            const { default: open } = await Promise.resolve().then(() => __importStar(require('open')));
            open(authURL);
            console.log(`\nIf the browser didn't open, visit:\n${authURL}\n`);
        });
        srv.on('error', reject);
    });
    const token = await exchangeCode(code, verifier);
    saveToken(token);
    console.log('Logged in successfully.');
}
