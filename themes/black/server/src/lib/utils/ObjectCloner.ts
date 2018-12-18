
export class ObjectCloner {
  static clone(obj: any): any {
    return JSON.parse(JSON.stringify(obj));
  }
}