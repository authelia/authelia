
export default function<T>(fn: () => Promise<T>, expectedErrorType: any = Error, logs = false): void {
  fn().then(() => {
    throw new Error("Should reject");
  }, (err: Error) => {
    if (!(err instanceof expectedErrorType)) {
      throw new Error(`Received error ${typeof err} != Expected error ${expectedErrorType}`);
    }
    if (logs) console.error(err);
  });
}