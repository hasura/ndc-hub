/**
 * @readonly Exposes the function as an NDC function (the function should only query data without making modifications)
 */
export function helloQuery(name?: string) {
  return `hello ${name ?? "world query"}`;
}

export function helloMutation(name?: string) {
  return `hello ${name ?? "world mutation"}`;
}