import axios from "axios";

import { ScopeDescription } from "@components/OpenIDConnect";
import { FlowID, UserCode } from "@constants/SearchParams";
import { OpenIDConnectConsentPath, OpenIDConnectDeviceAuthorizationPath } from "@services/Api";
import { Get, Post } from "@services/Client";

interface ConsentPostRequestBody {
    flow_id?: string;
    client_id: string;
    consent: boolean;
    pre_configure: boolean;
    claims?: string[];
    subflow?: string;
    user_code?: string;
}

interface ConsentPostResponseBody {
    redirect_uri?: string;
    flow_id?: string;
}

export interface ConsentGetResponseBody {
    client_id: string;
    client_description: string;
    scopes: string[];
    audience: string[];
    pre_configuration: boolean;
    claims: null | string[];
    essential_claims: null | string[];
    require_login: boolean;
}

export function getConsentResponse(flowID?: string, userCode?: string) {
    const params = new URLSearchParams();

    if (flowID) {
        params.append(FlowID, flowID);
    }

    if (userCode) {
        params.append(UserCode, userCode);
    }

    return Get<ConsentGetResponseBody>(`${OpenIDConnectConsentPath}?${params.toString()}`);
}

export function postConsentResponseAccept(
    preConfigure: boolean,
    clientID: string,
    claims: string[],
    flowID?: string,
    subflow?: string,
    userCode?: string,
) {
    const body: ConsentPostRequestBody = {
        claims: claims,
        client_id: clientID,
        consent: true,
        flow_id: flowID,
        pre_configure: preConfigure,
        subflow: subflow,
        user_code: userCode,
    };

    return Post<ConsentPostResponseBody>(OpenIDConnectConsentPath, body);
}

export function putDeviceCodeFlowUserCode(flowID: string, userCode: string) {
    const params = new URLSearchParams();

    params.append(FlowID, flowID);
    params.append(UserCode, userCode);

    return axios.put(OpenIDConnectDeviceAuthorizationPath, params);
}

export function postConsentResponseReject(clientID: string, flowID?: string, subflow?: string, userCode?: string) {
    const body: ConsentPostRequestBody = {
        client_id: clientID,
        consent: false,
        flow_id: flowID,
        pre_configure: false,
        subflow: subflow,
        user_code: userCode,
    };

    return Post<ConsentPostResponseBody>(OpenIDConnectConsentPath, body);
}

export function formatScope(scope: string, fallback: string): string {
    if (!scope.startsWith("scopes.") && scope !== "") {
        return scope;
    } else {
        return ScopeDescription(fallback);
    }
}

export function formatClaim(claim: string, fallback: string): string {
    if (!claim.startsWith("claims.") && claim !== "") {
        return claim;
    } else {
        return getClaimDescription(fallback);
    }
}

export function getClaimDescription(claim: string): string {
    switch (claim) {
        case "name":
            return "Display Name";
        case "sub":
            return "Unique Identifier";
        case "zoneinfo":
            return "Timezone";
        case "locale":
            return "Locale / Language";
        case "updated_at":
            return "Information Updated Time";
        case "profile":
        case "website":
        case "picture":
            return `${setClaimCase(claim)} URL`;
        default:
            return setClaimCase(claim);
    }
}

function setClaimCase(claim: string): string {
    claim = (claim.charAt(0).toUpperCase() + claim.slice(1)).replace("_verified", " (Verified)").replace("_", " ");

    let result = "";
    for (let i = 0; i < claim.length; i++) {
        if (i === 0 || claim[i - 1] === " ") {
            result += claim[i].toUpperCase();
        } else {
            result += claim[i];
        }
    }

    return result;
}
