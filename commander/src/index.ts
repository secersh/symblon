#!/usr/bin/env node
import { login, logout } from './auth';
import { publish } from './publish';

const [, , cmd, ...args] = process.argv;

async function main() {
  switch (cmd) {
    case 'login':
      await login();
      break;

    case 'logout':
      logout();
      break;

    case 'publish':
      if (!args[0]) {
        console.error('Usage: symblon publish <path>');
        process.exit(1);
      }
      await publish(args[0]);
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
      if (cmd) process.exit(1);
  }
}

main().catch((e) => {
  console.error(e.message ?? e);
  process.exit(1);
});
