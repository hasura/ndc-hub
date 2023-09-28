import { ErrorResponse } from "../schemas/ErrorResponse";

// ref: https://hasura.github.io/ndc-spec/specification/error-handling.html
type ErrorCode =
  | /** Bad Request - The request did not match the data connector's expectation based on this specification. */ 400
  | /** Forbidden - The request could not be handled because a permission check failed - for example, a mutation might fail because a check constraint was not met.*/ 403
  | /** Conflict - The request could not be handled because it would create a conflicting state for the data source - for example, a mutation might fail because a foreign key constraint was not met.*/ 409
  | /** Internal Server Error - The request could not be handled because of an error on the server */ 500
  | /** Not Supported - The request could not be handled because it relies on an unsupported capability. Note: this ought to indicate an error on the caller side, since the caller should not generate requests which are incompatible with the indicated capabilities. */ 501;

export class ConnectorError extends Error {
  statusCode: ErrorCode;
  message: ErrorResponse["message"];
  details?: ErrorResponse["details"];

  constructor(
    statusCode: ErrorCode,
    message: ErrorResponse["message"],
    details?: ErrorResponse["details"]
  ) {
    super(message);

    this.statusCode = statusCode;

    if (details !== undefined) {
      this.details = details;
    }

    return this;
  }
}

/** The request did not match the data connector's expectation based on this specification */
export class BadRequest extends ConnectorError {
  constructor(
    message: ErrorResponse["message"],
    details?: ErrorResponse["details"]
  ) {
    super(400, message, details);
  }
}

/** The request could not be handled because a permission check failed - for example, a mutation might fail because a check constraint was not met */
export class Forbidden extends ConnectorError {
  constructor(
    message: ErrorResponse["message"],
    details?: ErrorResponse["details"]
  ) {
    super(403, message, details);
  }
}

/** The request could not be handled because it would create a conflicting state for the data source - for example, a mutation might fail because a foreign key constraint was not met */
export class Conflict extends ConnectorError {
  constructor(
    message: ErrorResponse["message"],
    details?: ErrorResponse["details"]
  ) {
    super(409, message, details);
  }
}
/** The request could not be handled because of an error on the server */
export class InternalServerError extends ConnectorError {
  constructor(
    message: ErrorResponse["message"],
    details?: ErrorResponse["details"]
  ) {
    super(500, message, details);
  }
}
/** The request could not be handled because it relies on an unsupported capability. Note: this ought to indicate an error on the caller side, since the caller should not generate requests which are incompatible with the indicated capabilities */
export class NotSupported extends ConnectorError {
  constructor(
    message: ErrorResponse["message"],
    details?: ErrorResponse["details"]
  ) {
    super(501, message, details);
  }
}
