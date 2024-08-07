## Overview

The NodeJS Lambda connector allows you to expose TypeScript functions as Commands in your Hasura DDN Supergraph.

Here's an example function that defines and uses two object types. It is exported as a `@readonly` function so that it can be exposed from the Supergraph's GraphQL query schema.

```typescript
type FullName = {
  title: string
  firstName: string
  surname: string
}

type Greeting = {
  polite: string
  casual: string
}

/** @readonly */
export function greet(name: FullName): Greeting {
  return {
    polite: `Hello ${name.title} ${name.surname}`,
    casual: `G'day ${name.firstName}`
  }
}
```

The NodeJS Lambda connector introspects the TypeScript types you use on your exported functions to determine the schema to use in your Supergraph. You are able to import any NodeJS npm package and use it.

The NodeJS Lambda connector enables you to:

* Add business logic to your Supergraph by writing normal NodeJS TypeScript code
* Expose data from any external API by writing TypeScript code that invokes the API and returns the data
* Implement complex database logic manually by connecting to your database via a native driver library imported from npm

To learn more about this connector and how to use it, please see the [Hasura DDN Documentation](https://hasura.io/docs/3.0/connectors/typescript).

## More Information
* [Hasura DDN Documentation](https://hasura.io/docs/3.0/connectors/typescript)
* [ndc-nodejs-lambda GitHub Repository](https://github.com/hasura/ndc-nodejs-lambda)