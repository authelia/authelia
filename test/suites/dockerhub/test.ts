import AutheliaSuite from '../../helpers/context/AutheliaSuite';
import DockerCompose from '../../helpers/context/DockerCompose';
import { composeFiles } from './environment';
import Assert from 'assert';
import SimpleAuthentication from './scenarii/SimpleAuthentication';

AutheliaSuite('Dockerhub', __dirname, function() {
  this.timeout(15000);
  const dockerCompose = new DockerCompose(composeFiles);
  
  describe('Check the container', function() {
    it('should be running', async function() {
      const stdout = await dockerCompose.ps();
      const lines = stdout.split("\n");
      const autheliaLine = lines.filter(l => l.indexOf('authelia_1') > -1);
      if (autheliaLine.length != 1) {
        throw new Error('Authelia container not found...');
      }
      // check if the container is up.
      Assert(autheliaLine[0].indexOf(' Up ') > -1);
    });
  });

  describe.only('Simple authentication', SimpleAuthentication);
});