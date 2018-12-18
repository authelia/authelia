
declare type Handler = () => void;

export interface Handlers {
  onFadedIn: Handler;
  onFadedOut: Handler;
}

export interface INotifier {
  success(msg: string, handlers?: Handlers): void;
  error(msg: string, handlers?: Handlers): void;
  warning(msg: string, handlers?: Handlers): void;
  info(msg: string, handlers?: Handlers): void;
}