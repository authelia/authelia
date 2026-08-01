// generateRandomPassword
export function generateRandomPassword(length: number) {
    const lowercase = "abcdefghijklmnopqrstuvwxyz";
    const uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ";
    const numbers = "0123456789";
    const special = "!@#$%^&*_-+=";

    // Helper function to get cryptographically secure random integer
    const getSecureRandom = (max: number): number => {
        const randomBuffer = new Uint32Array(1);
        crypto.getRandomValues(randomBuffer);
        const maxMultiple = Math.floor((2 ** 32 - 1) / max) * max;
        if (randomBuffer[0] >= maxMultiple) {
            return getSecureRandom(max);
        }
        return randomBuffer[0] % max;
    };

    const getSecureRandomChar = (str: string): string => {
        return str[getSecureRandom(str.length)];
    };
    const guaranteedChars = [
        getSecureRandomChar(lowercase),
        getSecureRandomChar(uppercase),
        getSecureRandomChar(numbers),
        getSecureRandomChar(special),
    ];

    const allChars = lowercase + uppercase + numbers + special;
    const remainingLength = length - guaranteedChars.length;
    const remainingChars = Array.from({ length: remainingLength }, () => getSecureRandomChar(allChars));

    const allCharsArray = [...guaranteedChars, ...remainingChars];
    for (let i = allCharsArray.length - 1; i > 0; i--) {
        const j = getSecureRandom(i + 1);
        [allCharsArray[i], allCharsArray[j]] = [allCharsArray[j], allCharsArray[i]];
    }

    return allCharsArray.join("");
}
