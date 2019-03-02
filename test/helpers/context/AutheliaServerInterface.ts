
interface AutheliaServerInterface {
  start(): Promise<void>;
  stop(): Promise<void> 
}

export default AutheliaServerInterface;