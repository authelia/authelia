package middlewares

// InternalError is the error message sent when there was an internal error but it should
// be hidden to the end user. In that case the error should be in the server logs.
const InternalError = "Internal error."

// UnauthorizedError is the error message sent when the user is not authorized.
const UnauthorizedError = "You're not authorized."
