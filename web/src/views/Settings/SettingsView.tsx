import { Box, Typography } from "@mui/material";

import SettingsLayout from "@layouts/SettingsLayout";

export interface Props {}

const SettingsView = function (props: Props) {
    return (
        <SettingsLayout>
            <Box>
                <Typography>Placeholder</Typography>
            </Box>
        </SettingsLayout>
    );
};

export default SettingsView;
