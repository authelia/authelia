
// returns true for 1 or more matches, where 'a' is an array and 'b' is a search string or an array of multiple search strings
function contains(a: string, character: string) {
  // string match
  return a.indexOf(character) > -1;
}

function containsOneOf(s: string, characters: string[]) {
  return characters
    .map((character: string) => { return contains(s, character); })
    .reduce((acc: boolean, current: boolean) => { return acc || current; }, false);
}

export class Sanitizer {
  static sanitize(input: string): string {
    const forbiddenChars = [",", "\\", "'", "#", "+", "<", ">", ";", "\"", "="];
    if (containsOneOf(input, forbiddenChars)) {
      throw new Error("Input containing unsafe characters.");
    }

    if (input != input.trim()) {
      throw new Error("Input has unexpected spaces.");
    }

    return input;
  }
}
