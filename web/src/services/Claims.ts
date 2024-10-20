export function setClaimCase(claim: string): string {
    claim.charAt(0).toUpperCase();
    claim.replace("_verified", " (Verified)");
    claim.replace("_", " ");

    for (let i = 0; i < claim.length; i++) {
        const j = i + 1;

        if (claim[i] === " " && j < claim.length) {
            claim.charAt(j).toUpperCase();
        }
    }
    return claim;
}
