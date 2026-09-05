import { Fragment, useCallback, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { Spinner } from "@components/UI/Spinner";
import { useNotifications } from "@contexts/NotificationsContext";
import { useRemoteCall } from "@hooks/RemoteCall";
import { useUserInfoPOST } from "@hooks/UserInfo";
import {
    deleteOpenIDConnectLinkPending,
    getOpenIDConnectLinks,
    getOpenIDConnectProviders,
    postOpenIDConnectStart,
    putOpenIDConnectLink,
} from "@services/OpenIDConnectRelyingParty";
import { UserSessionElevation, getUserSessionElevation } from "@services/UserSessionElevation";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";
import SecondFactorDialog from "@views/Settings/Common/SecondFactorDialog";
import OpenIDConnectLinkDeleteDialog from "@views/Settings/OpenIDConnect/OpenIDConnectLinkDeleteDialog";
import OpenIDConnectLinkPendingPanel from "@views/Settings/OpenIDConnect/OpenIDConnectLinkPendingPanel";
import OpenIDConnectLinksPanel from "@views/Settings/OpenIDConnect/OpenIDConnectLinksPanel";

const OpenIDConnectView = function () {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification, createSuccessNotification } = useNotifications();

    const [userInfo, fetchUserInfo, , fetchUserInfoError] = useUserInfoPOST();
    const [links, fetchLinks, , fetchLinksError] = useRemoteCall(getOpenIDConnectLinks);
    const [providers, fetchProviders] = useRemoteCall(getOpenIDConnectProviders);

    const [refreshState, setRefreshState] = useState(0);
    const [linking, setLinking] = useState<string>();

    const [elevation, setElevation] = useState<UserSessionElevation>();
    const [dialogSFOpening, setDialogSFOpening] = useState(false);
    const [dialogIVOpening, setDialogIVOpening] = useState(false);

    const [dialogAcceptOpening, setDialogAcceptOpening] = useState(false);

    const [dialogDeleteOpen, setDialogDeleteOpen] = useState(false);
    const [dialogDeleteOpening, setDialogDeleteOpening] = useState(false);
    const [deleteLinkID, setDeleteLinkID] = useState<number>();

    const handleRefreshState = useCallback(() => {
        setRefreshState((refreshState) => refreshState + 1);
    }, []);

    useEffect(() => {
        fetchUserInfo();
    }, [fetchUserInfo]);

    useEffect(() => {
        fetchLinks();
    }, [fetchLinks, refreshState]);

    useEffect(() => {
        fetchProviders();
    }, [fetchProviders]);

    useEffect(() => {
        if (fetchUserInfoError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", { item: translate("user preferences") }),
            );
        }
    }, [fetchUserInfoError, createErrorNotification, translate]);

    useEffect(() => {
        if (fetchLinksError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", { item: translate("Linked Accounts") }),
            );
        }
    }, [fetchLinksError, createErrorNotification, translate]);

    const handleResetStateOpening = () => {
        setDialogSFOpening(false);
        setDialogIVOpening(false);
        setDialogAcceptOpening(false);
        setDialogDeleteOpening(false);
    };

    const handleResetState = useCallback(() => {
        handleResetStateOpening();

        setElevation(undefined);

        setDialogDeleteOpen(false);
        setDeleteLinkID(undefined);
    }, []);

    const handleElevationRefresh = async () => {
        const result = await getUserSessionElevation();

        setElevation(result);
        return result;
    };

    const handleElevation = () => {
        handleElevationRefresh().catch(console.error);

        setDialogSFOpening(true);
    };

    const notifyFailure = useCallback(
        (action: string, providerName: string) => {
            createErrorNotification(
                translate("There was a problem {{action}} the {{item}}", {
                    action: translate(action),
                    item: `${providerName} account`,
                }),
            );
        },
        [createErrorNotification, translate],
    );

    const notifySuccess = useCallback(
        (action: string, providerName: string) => {
            createSuccessNotification(
                translate("Successfully {{action}} the {{item}}", {
                    action: translate(action),
                    item: `${providerName} account`,
                }),
            );
        },
        [createSuccessNotification, translate],
    );

    const handleAcceptConfirmed = useCallback(() => {
        const providerName = links?.pending?.provider_name ?? "";

        (async () => {
            try {
                await putOpenIDConnectLink();
            } catch (err) {
                console.error(err);
                notifyFailure("linking", providerName);
                handleResetState();

                return;
            }

            notifySuccess("linked", providerName);

            handleResetState();
            handleRefreshState();
        })();
    }, [links, handleRefreshState, handleResetState, notifyFailure, notifySuccess]);

    const handleOpenDialogDelete = useCallback(() => {
        handleResetStateOpening();
        setDialogDeleteOpen(true);
    }, []);

    const handleSFDialogClosed = (ok: boolean, changed: boolean) => {
        if (!ok) {
            console.warn("Second Factor dialog close callback failed, it was likely cancelled by the user.");

            handleResetState();

            return;
        }

        if (changed) {
            handleElevationRefresh()
                .catch(console.error)
                .then((refreshedElevation) => {
                    if (refreshedElevation) {
                        const isElevatedFromRefresh =
                            refreshedElevation.elevated || refreshedElevation.skip_second_factor;
                        if (isElevatedFromRefresh) {
                            setElevation(undefined);
                            if (dialogAcceptOpening) {
                                handleAcceptConfirmed();
                            } else if (dialogDeleteOpening) {
                                handleOpenDialogDelete();
                            }
                        } else {
                            setDialogIVOpening(true);
                        }
                    }
                });
        } else {
            const isElevated = elevation && (elevation.elevated || elevation.skip_second_factor);
            if (isElevated) {
                setElevation(undefined);
                if (dialogAcceptOpening) {
                    handleAcceptConfirmed();
                } else if (dialogDeleteOpening) {
                    handleOpenDialogDelete();
                }
            } else {
                setDialogIVOpening(true);
            }
        }
    };

    const handleSFDialogOpened = () => {
        setDialogSFOpening(false);
    };

    const handleIVDialogClosed = useCallback(
        (ok: boolean) => {
            if (!ok) {
                console.warn(
                    "Identity Verification dialog close callback failed, it was likely cancelled by the user.",
                );

                handleResetState();

                return;
            }

            setElevation(undefined);

            if (dialogAcceptOpening) {
                handleAcceptConfirmed();
            } else if (dialogDeleteOpening) {
                handleOpenDialogDelete();
            }
        },
        [handleResetState, handleAcceptConfirmed, handleOpenDialogDelete, dialogAcceptOpening, dialogDeleteOpening],
    );

    const handleIVDialogOpened = useCallback(() => {
        setDialogIVOpening(false);
    }, []);

    const handleAccept = () => {
        setDialogAcceptOpening(true);
        handleElevation();
    };

    const handleDecline = () => {
        const providerName = links?.pending?.provider_name ?? "";

        (async () => {
            try {
                await deleteOpenIDConnectLinkPending();
            } catch (err) {
                console.error(err);
                notifyFailure("declining", providerName);

                return;
            }

            notifySuccess("declined", providerName);

            handleRefreshState();
        })();
    };

    // The linking flow is performed again in the users own browser rather than carried across a session boundary. The
    // provider identifier is validated by the start endpoint which rejects providers which are not configured.
    const handleLink = useCallback(
        (id: string) => {
            const providerName = providers?.find((provider) => provider.id === id)?.name ?? id;

            setLinking(id);

            (async () => {
                try {
                    const response = await postOpenIDConnectStart(id, { keepMeLoggedIn: false });

                    window.location.assign(response.authorization_url);
                } catch (err) {
                    console.error(err);
                    notifyFailure("linking", providerName);
                    setLinking(undefined);
                }
            })();
        },
        [providers, notifyFailure],
    );

    const handleDelete = (id: number) => {
        setDeleteLinkID(id);
        setDialogDeleteOpening(true);
        handleElevation();
    };

    const deleteLink = links?.links.find((link) => link.id === deleteLinkID);

    // A provider the user already has a link for is not offered again. The backend permits a single link per provider
    // per user so a second attempt would be a guaranteed conflict.
    const providersUnlinked = (providers ?? []).filter(
        (provider) =>
            !links?.links.some((link) => link.provider === provider.id) && links?.pending?.provider !== provider.id,
    );

    return (
        <Fragment>
            <SecondFactorDialog
                info={userInfo}
                elevation={elevation}
                opening={dialogSFOpening}
                handleClosed={handleSFDialogClosed}
                handleOpened={handleSFDialogOpened}
            />
            <IdentityVerificationDialog
                elevation={elevation}
                opening={dialogIVOpening}
                handleClosed={handleIVDialogClosed}
                handleOpened={handleIVDialogOpened}
            />
            <OpenIDConnectLinkDeleteDialog
                open={dialogDeleteOpen}
                link={deleteLink}
                handleClose={() => {
                    handleResetState();
                    handleRefreshState();
                }}
            />
            <div className="grid grid-cols-1 gap-4">
                {links?.pending ? (
                    <div className="w-full">
                        <OpenIDConnectLinkPendingPanel
                            pending={links.pending}
                            onAccept={handleAccept}
                            onDecline={handleDecline}
                        />
                    </div>
                ) : null}
                <div className="w-full">
                    {links === undefined ? (
                        <Spinner size={20} />
                    ) : (
                        <OpenIDConnectLinksPanel
                            links={links.links}
                            providers={providersUnlinked}
                            linking={linking}
                            onDelete={handleDelete}
                            onLink={handleLink}
                        />
                    )}
                </div>
            </div>
        </Fragment>
    );
};

export default OpenIDConnectView;
