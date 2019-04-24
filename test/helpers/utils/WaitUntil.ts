import sleep from "./sleep";

export default function WaitUntil(
  fn: () => Promise<boolean>, timeout: number = 15000,
  interval: number = 1000, waitBefore: number = 0, waitAfter: number = 0): Promise<void> {

  return new Promise(async (resolve, reject) => {
    const timer = setTimeout(() => { throw new Error('Timeout') }, timeout);
    if (waitBefore > 0)
      await sleep(waitBefore);
    while (true) {
      try {
        const res = await fn();
        if (res && res === true) {
          clearTimeout(timer);
          break;
        }
        await sleep(interval);
      } catch (err) {
        console.error(err);
        reject(err);
        return;
      }
    }
    
    if (waitAfter > 0)
      await sleep(waitAfter);
    resolve();
  });
}