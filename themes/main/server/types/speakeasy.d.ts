declare module "speakeasy" {
  export = speakeasy

  interface SharedOptions {
    encoding?: string
    algorithm?: string
  }

  interface DigestOptions extends SharedOptions {
    secret: string
    counter: number
  }

  interface HOTPOptions extends SharedOptions {
    secret: string
    counter: number
    digest?: Buffer
    digits?: number
  }

  interface HOTPVerifyOptions extends SharedOptions {
    secret: string
    token: string
    counter: number
    digits?: number
    window?: number
  }

  interface TOTPOptions extends SharedOptions {
    secret: string
    time?: number
    step?: number
    epoch?: number
    counter?: number
    digits?: number
  }

  interface TOTPVerifyOptions extends SharedOptions {
    secret: string
    token: string
    time?: number
    step?: number
    epoch?: number
    counter?: number
    digits?: number
    window?: number
  }

  interface GenerateSecretOptions {
    length?: number
    symbols?: boolean
    otpauth_url?: boolean
    name?: string
    issuer?: string
  }

  interface GeneratedSecret {
    ascii: string
    hex: string
    base32: string
    qr_code_ascii: string
    qr_code_hex: string
    qr_code_base32: string
    google_auth_qr: string
    otpauth_url: string
  }

  interface OTPAuthURLOptions extends SharedOptions {
    secret: string
    label: string
    type?: string
    counter?: number
    issuer?: string
    digits?: number
    period?: number
  }

  interface Speakeasy {
    digest: (options: DigestOptions) => Buffer
    hotp: {
      (options: HOTPOptions): string,
      verifyDelta: (options: HOTPVerifyOptions) => boolean,
      verify: (options: HOTPVerifyOptions) => boolean,
    }
    totp: {
      (options: TOTPOptions): string
      verifyDelta: (options: TOTPVerifyOptions) => boolean,
      verify: (options: TOTPVerifyOptions) => boolean,
    }
    generateSecret: (options?: GenerateSecretOptions) => GeneratedSecret
    generateSecretASCII: (length?: number, symbols?: boolean) => string
    otpauthURL: (options: OTPAuthURLOptions) => string
  }

  const speakeasy: Speakeasy
}