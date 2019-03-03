

// Error thrown when the authentication failed when checking
// user/password.
class AuthenticationError extends Error {
  constructor(msg: string) {
    super(msg);
  }
}

export default AuthenticationError;