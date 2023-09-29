
import {start} from 'https://raw.githubusercontent.com/hasura/ndc-sdk-typescript/main/src/server.ts';
import {Connector} from 'https://raw.githubusercontent.com/hasura/ndc-sdk-typescript/main/src/connector.ts';

console.log('hello world');

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