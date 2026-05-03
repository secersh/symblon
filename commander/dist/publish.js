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
exports.publish = publish;
const fs = __importStar(require("fs"));
const path = __importStar(require("path"));
const archiver = __importStar(require("archiver"));
const auth_1 = require("./auth");
const DEFAULT_REGISTRAR = 'https://api.symblon.cc';
function findAgentDir(p) {
    if (fs.existsSync(path.join(p, 'agent.yaml')))
        return p;
    for (const entry of fs.readdirSync(p, { withFileTypes: true })) {
        if (entry.isDirectory()) {
            const candidate = path.join(p, entry.name);
            if (fs.existsSync(path.join(candidate, 'agent.yaml')))
                return candidate;
        }
    }
    throw new Error(`No agent.yaml found in ${p}`);
}
function zipDir(dir) {
    return new Promise((resolve, reject) => {
        const chunks = [];
        const archive = archiver.default('zip');
        archive.on('data', (chunk) => chunks.push(chunk));
        archive.on('end', () => resolve(Buffer.concat(chunks)));
        archive.on('error', reject);
        // Include the agent dir with its folder name as the zip root
        archive.directory(dir, path.basename(dir));
        archive.finalize();
    });
}
async function publish(inputPath) {
    const token = (0, auth_1.loadToken)();
    const registrar = process.env.REGISTRAR_URL ?? DEFAULT_REGISTRAR;
    const agentDir = findAgentDir(path.resolve(inputPath));
    console.log(`Publishing ${path.basename(agentDir)}...`);
    const zip = await zipDir(agentDir);
    const form = new FormData();
    const ab = zip.buffer.slice(zip.byteOffset, zip.byteOffset + zip.byteLength);
    form.append('package', new Blob([ab], { type: 'application/zip' }), `${path.basename(agentDir)}.zip`);
    const res = await fetch(`${registrar}/registrar/v1/agents`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${token.access_token}` },
        body: form,
    });
    const body = await res.json().catch(() => ({}));
    if (res.status === 201) {
        console.log(`Published: ${body.ref}`);
    }
    else if (res.status === 409) {
        console.log('Already exists — skipping (bump version to republish).');
    }
    else {
        console.error(`Error (HTTP ${res.status}):`, body);
        process.exit(1);
    }
}
