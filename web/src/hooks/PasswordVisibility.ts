// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useCallback, useState } from "react";

export const usePasswordVisibility = () => {
    const [showPassword, setShowPassword] = useState(false);

    const toggle = useCallback(() => setShowPassword((prev) => !prev), []);

    const toggleProps = {
        "aria-pressed": showPassword,
        onClick: toggle,
    };

    return {
        showPassword,
        toggleProps,
    };
};

export default usePasswordVisibility;
