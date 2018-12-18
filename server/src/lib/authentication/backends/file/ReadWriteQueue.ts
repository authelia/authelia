import Fs = require("fs");

type Callback = (err: Error, data?: string) => void;
type ContentAndCallback = [string, Callback] | [string, string, Callback];

/**
 * WriteQueue is a queue synchronizing writes to a file.
 *
 * Example of use:
 *
 * queue.add(mycontent, (err) => {
 *    // do whatever you want here.
 *    queue.next();
 * })
 */
export class ReadWriteQueue {
  private filePath: string;
  private queue: ContentAndCallback[];

  constructor (filePath: string) {
    this.queue = [];
    this.filePath = filePath;
  }

  next () {
    if (this.queue.length === 0)
      return;

    const task = this.queue[0];

    if (task[0] == "write") {
      Fs.writeFile(this.filePath, task[1], "utf-8", (err) => {
        this.queue.shift();
        const cb = task[2] as Callback;
        cb(err);
      });
    }
    else if (task[0] == "read") {
      Fs.readFile(this.filePath, { encoding: "utf-8"} , (err, data) => {
        this.queue.shift();
        const cb = task[1] as Callback;
        cb(err, data);
      });
    }
  }

  write (content: string, cb: Callback) {
    this.queue.push(["write", content, cb]);
    if (this.queue.length === 1) {
      this.next();
    }
  }

  read (cb: Callback) {
    this.queue.push(["read", cb]);
    if (this.queue.length === 1) {
      this.next();
    }
  }
}