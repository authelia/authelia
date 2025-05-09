import { render } from '@react-email/render';
import * as React from "react";
import * as fs from "node:fs";

import Event from './emails/Event';
import IdentityVerificationJWT from "./emails/IdentityVerificationJWT";
import IdentityVerificationOTC from "./emails/IdentityVerificationOTC";
import NewLogin from "./emails/NewLogin";

const optsHTML = {
	pretty: false,
	plainText: false,
};

const optsTXT = {
	pretty: false,
	plainText: true,
};

	/*
		Generate Event Email
	*/

async function doRender() {
	const propsEvent = {
		title: "{{ .Title }}",
		displayName: "{{ .DisplayName }}",
        bodyPrefix: "{{ .BodyPrefix }}",
        bodyEvent: "{{ .BodyEvent }}",
        bodySuffix: "{{ .BodySuffix }}",
		remoteIP: "{{ .RemoteIP }}",
		detailsKey: "{{ $key }}",
		detailsValue: "{{ index $.Details $key }}",
		detailsPrefix: "{{- $keys := sortAlpha (keys .Details) }}{{- range $key := $keys }}",
		detailsSuffix: "{{ end }}",
	};

	fs.writeFileSync('../embed/notification/Event.html', await render(<Event {...propsEvent} />, optsHTML));
	fs.writeFileSync('../embed/notification/Event.txt', await render(<Event {...propsEvent} />, optsTXT));

	const propsEventNoPreview = {
		...propsEvent,
		hidePreview: true,
	};

	fs.writeFileSync('../../../examples/templates/notifications/no-preview/Event.html', await render(<Event {...propsEventNoPreview} />, optsHTML));


	/*
		Generate New Login Email
 	*/

	const propsNewLogin = {
		title: "{{ .Title }}",
		domain: "{{ .Domain }}",
		date: "{{ .Date }}",
		userAgent: "{{ .UserAgent}}",
		displayName: "{{ .DisplayName }}",
		remoteIP: "{{ .RemoteIP }}",
	};

	fs.writeFileSync('../embed/notification/NewLogin.html', await render(<NewLogin {...propsNewLogin} />, optsHTML));
	fs.writeFileSync('../embed/notification/NewLogin.txt', await render(<NewLogin {...propsNewLogin} />, optsTXT));

	const propsNewLoginNoPreview = {
		...propsNewLogin,
		hidePreview: true,
	};

	fs.writeFileSync('../../../examples/templates/notifications/no-preview/NewLogin.html', await render(<NewLogin {...propsNewLoginNoPreview} />, optsHTML));

	/*
		Generate Identity Verification JWT Email
	*/

	const propsJWT = {
		title: "{{ .Title }}",
		displayName: "{{ .DisplayName }}",
        domain: "{{ .Domain }}",
		remoteIP: "{{ .RemoteIP }}",
		link: "{{ .LinkURL }}",
		linkText: "{{ .LinkText }}",
		revocationLinkURL: "{{ .RevocationLinkURL }}",
		revocationLinkText: "{{ .RevocationLinkText }}",
	};

	fs.writeFileSync('../embed/notification/IdentityVerificationJWT.html', await render(<IdentityVerificationJWT {...propsJWT} />, optsHTML));
	fs.writeFileSync('../embed/notification/IdentityVerificationJWT.txt', await render(<IdentityVerificationJWT {...propsJWT} />, optsTXT));

	const propsJWTNoPreview = {
		...propsJWT,
		hidePreview: true,
	};

	fs.writeFileSync('../../../examples/templates/notifications/no-preview/IdentityVerificationJWT.html', await render(<IdentityVerificationJWT {...propsJWTNoPreview} />, optsHTML));

	/*
		Generate Identity Verification OTP Email
	*/
	const propsOTC = {
		title: "{{ .Title }}",
		displayName: "{{ .DisplayName }}",
        domain: "{{ .Domain }}",
		remoteIP: "{{ .RemoteIP }}",
		oneTimeCode: "{{ .OneTimeCode }}",
		revocationLinkURL: "{{ .RevocationLinkURL }}",
		revocationLinkText: "{{ .RevocationLinkText }}",
	};

	fs.writeFileSync('../embed/notification/IdentityVerificationOTC.html', await render(<IdentityVerificationOTC {...propsOTC} />, optsHTML));
	fs.writeFileSync('../embed/notification/IdentityVerificationOTC.txt', await render(<IdentityVerificationOTC {...propsOTC} />, optsTXT));

	const propsOTCNoPreview = {
		...propsOTC,
		hidePreview: true,
	};

	fs.writeFileSync('../../../examples/templates/notifications/no-preview/IdentityVerificationOTC.html', await render(<IdentityVerificationOTC {...propsOTCNoPreview} />, optsHTML));
}

doRender().then();


