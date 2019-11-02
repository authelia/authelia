import AutheliaSuite from "../../helpers/context/AutheliaSuite";
import { exec } from '../../helpers/utils/exec';
import NetworkACLs from "./scenarii/NetworkACLs";

AutheliaSuite(__dirname, function () {
    this.timeout(10000);

    beforeEach(async function () {
        await exec(`cp ${__dirname}/../../../suites/NetworkACL/users.yml /tmp/authelia/suites/NetworkACL/users.yml`);
    });

    describe("Network ACLs", NetworkACLs);
});