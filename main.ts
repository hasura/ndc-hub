
/**
 * TODO: 
 * 
 * Create PR
 * Share SDK issues with Benoit
 * Resolve import errors for Deno (via import map?) for github.com/hasura/ndc-sdk-typescript
 * Convert server.ts to connector protocol
 * Remove rust harness
 * Update docker to leverage deno implementation
 * Do start-time inference on functions
 * Have schema cache respecting flag --schema: /schema.json - Creates if missing, uses if present
 * Manage src locations better
 * CMD parsing library
 * Subprocess library
 * File-reading library
 * Have local dev supported by `deno --watch`
 * 
 */


/**
 * Importing TS SDK Dependency
 * https://github.com/hasura/ndc-sdk-typescript/tree/main
 * 
 * Currently not working (due to missing import map?)
 */

// import {start} from 'https://raw.githubusercontent.com/hasura/ndc-sdk-typescript/main/src/server.ts';
// import {Connector} from 'https://raw.githubusercontent.com/hasura/ndc-sdk-typescript/main/src/connector.ts';

/**
 * Subprocesses:
 * https://docs.deno.com/runtime/tutorials/subprocess
 */

const command = new Deno.Command(Deno.execPath(), {
  args: [
    "eval",
    "console.log('hello'); console.error('world')",
  ],
});

// create subprocess and collect output
const { code, stdout, stderr } = await command.output();

console.assert(code === 0);
console.assert("world\n" === new TextDecoder().decode(stderr));
console.log(new TextDecoder().decode(stdout));

/**
 * Command line arguments:
 * https://examples.deno.land/command-line-arguments
 * 
 * Or via `cmd` library
 * https://deno.land/x/cmd@v1.2.0#action-handler-subcommands
 */

import { program } from 'https://deno.land/x/cmd@v1.2.0'

program
  .command('rm <dir>')
  .option('-r, --recursive', 'Remove recursively')
  .action(function (dir, cmdObj) {
    console.log('remove ' + dir + (cmdObj.recursive ? ' recursively' : ''))
  })

// program.parse(process.argv)