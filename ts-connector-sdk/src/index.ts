import { Connector } from "./connector";
import { Command, Option, InvalidOptionArgumentError } from "commander";
import { start_server } from "./server";
import { start_configuration_server } from "./configuration_server";

export function start<Configuration, State>(
  connector: Connector<Configuration, State>
) {
  const program = new Command();

  program.addCommand(serve_command(connector));
  program.addCommand(configuration_command(connector));

  program.parseAsync(process.argv).catch(console.error);
}

function serve_command(connector) {
  return new Command("serve")
    .addOption(
      new Option("--configuration")
        .env("CONFIGURATION_FILE")
        .makeOptionMandatory(true)
    )
    .addOption(
      new Option("--port").env("PORT").default(8100).argParser(parseIntOption)
    )
    .addOption(new Option("--service-token-secret").env("SERVICE_TOKEN_SECRET"))
    .addOption(new Option("--otlp_endpoint").env("OTLP_ENDPOINT"))
    .addOption(new Option("--service-name").env("OTEL_SERVICE_NAME"))
    .action(async (options) => {
      await start_server(connector, options);
    });
}

function configuration_command(connector) {
  return new Command("configuration").addCommand(
    new Command("serve")
      .addOption(
        new Option("--port").env("PORT").default(9100).argParser(parseIntOption)
      )
      .action(async (options) => {
        await start_configuration_server(connector, options);
      })
  );
}

function parseIntOption(value, previous) {
  // parseInt takes a string and a radix
  const parsedValue = parseInt(value, 10);
  if (isNaN(parsedValue)) {
    throw new InvalidOptionArgumentError("Not a valid integer.");
  }
  return parsedValue;
}
