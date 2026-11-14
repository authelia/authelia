// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { UserLock } from "lucide-react";

export interface Props {}

const LockIcon = function () {
    return <UserLock className="lock-icon size-17.5 text-[oklch(from_var(--destructive)_0.55_c_h)]" />;
};

export default LockIcon;
