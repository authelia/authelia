import { getBasePath } from "@utils/BasePath";

__webpack_public_path__ = "/";

if (getBasePath() !== "") {
    __webpack_public_path__ = getBasePath() + "/";
}
