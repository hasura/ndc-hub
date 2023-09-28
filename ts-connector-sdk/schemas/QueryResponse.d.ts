/* eslint-disable */
/**
 * This file was automatically generated by json-schema-to-typescript.
 * DO NOT MODIFY IT BY HAND. Instead, modify the source JSONSchema file,
 * and run json-schema-to-typescript to regenerate this file.
 */

/**
 * Query responses may return multiple RowSets when using foreach queries Else, there should always be exactly one RowSet
 */
export type QueryResponse = RowSet[];

export interface RowSet {
  /**
   * The results of the aggregates returned by the query
   */
  aggregates?: {
    [k: string]: unknown;
  } | null;
  /**
   * The rows returned by the query, corresponding to the query's fields
   */
  rows?:
    | {
        [k: string]: RowFieldValue;
      }[]
    | null;
  [k: string]: unknown;
}
export interface RowFieldValue {
  [k: string]: unknown;
}