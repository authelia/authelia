import { useBasePath } from "./BasePath";

__webpack_public_path__ = "/"

if (useBasePath() !== "") {
  __webpack_public_path__ = useBasePath() + "/"
}