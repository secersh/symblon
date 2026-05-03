#!/usr/bin/env node
"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const auth_1 = require("./auth");
const publish_1 = require("./publish");
const [, , cmd, ...args] = process.argv;
async function main() {
    switch (cmd) {
        case 'login':
            await (0, auth_1.login)();
            break;
        case 'logout':
            (0, auth_1.logout)();
            break;
        case 'publish':
            if (!args[0]) {
                console.error('Usage: symblon publish <path>');
                process.exit(1);
            }
            await (0, publish_1.publish)(args[0]);
            break;
        default:
            console.log(`symblon — agent management CLI

Commands:
  login              Authenticate via GitHub
  logout             Remove stored credentials
  publish <path>     Publish an agent package

Options:
  REGISTRAR_URL      Override registrar (default: https://api.symblon.cc)
`);
            if (cmd)
                process.exit(1);
    }
}
main().catch((e) => {
    console.error(e.message ?? e);
    process.exit(1);
});
