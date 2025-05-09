import { useCallback } from "react";

import { useSearchParams } from "react-router-dom";

import { LogoutRoute as SignOutRoute } from "@constants/Routes";
import { RedirectionRestoreURL, RedirectionURL } from "@constants/SearchParams";
import { useRouterNavigate } from "@hooks/RouterNavigate";

export function useSignOut() {
    const navigate = useRouterNavigate();
    const [query] = useSearchParams();

    return useCallback(
        (preserve: boolean) => {
            if (preserve) {
                if (!query.has(RedirectionURL)) {
                    navigate(SignOutRoute, preserve, preserve, preserve);
                } else {
                    const search = new URLSearchParams();

                    query.forEach((value, key) => {
                        if (key !== RedirectionURL) {
                            search.set(key, value);
                        } else {
                            search.set(RedirectionRestoreURL, value);
                        }
                    });

                    navigate(SignOutRoute, preserve, preserve, preserve, search);
                }
            } else {
                navigate(SignOutRoute, preserve, preserve, preserve);
            }
        },
        [navigate, query],
    );
}
