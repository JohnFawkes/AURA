// Function that takes supported data and returns singular/plural form of `word`
export function makePlural(data: unknown, word: string): string {
  if (Array.isArray(data)) {
    return data.length === 1 ? word : `${word}s`;
  }

  if (typeof data === "object" && data !== null) {
    return Object.keys(data).length === 1 ? word : `${word}s`;
  }

  if (typeof data === "number") {
    return data === 1 ? word : `${word}s`;
  }

  return `${word}s`;
}
