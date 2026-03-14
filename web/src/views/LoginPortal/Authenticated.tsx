import { Box, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

import SuccessIcon from "@components/SuccessIcon";

const Authenticated = function () {
    const { t: translate } = useTranslation();

    return (
        <Box id="authenticated-stage">
            <Box sx={{ flex: "0 0 100%", marginBottom: (theme) => theme.spacing(2) }}>
                <SuccessIcon />
            </Box>
            <Typography>{translate("Authenticated")}</Typography>
        </Box>
    );
};

export default Authenticated;
