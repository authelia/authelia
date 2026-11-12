// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { AnimatedLogo } from "./AnimatedLogo";

export interface Props {
    message: string;
}

const BaseLoadingPage = function ({ message }: Props) {
    return (
        <div className="grid min-h-screen items-center justify-center">
            <div className="inline-block text-center">
                <AnimatedLogo />

                <p className="m-4 font-bold text-lg">{message}...</p>
            </div>
        </div>
    );
};

export default BaseLoadingPage;
