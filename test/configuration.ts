import Bluebird = require("bluebird");
import YamlJS = require("yamljs");
import Fs = require("fs");
import ChildProcess = require("child_process");

const execAsync = Bluebird.promisify(ChildProcess.exec);

export class Configuration {
  private outputPath: string;

  setup(
    inputPath: string,
    outputPath: string,
    updateFn: (configuration: any) => void)
    : Bluebird<void> {

    console.log("[CONFIGURATION] setup");
    this.outputPath = outputPath;
    return new Bluebird((resolve, reject) => {
      const configuration = YamlJS.load(inputPath);
      updateFn(configuration);
      const configurationStr = YamlJS.stringify(configuration);
      Fs.writeFileSync(outputPath, configurationStr);
      resolve();
    });
  }

  cleanup(): Bluebird<{}> {
    console.log("[CONFIGURATION] cleanup");
    return execAsync(`rm ${this.outputPath}`);
  }
}