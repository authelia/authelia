import { render } from '@react-email/render';
import * as React from "react";
import * as fs from "node:fs";

import Event from './emails/Event';
import IdentityVerificationJWT from "./emails/IdentityVerificationJWT";
import IdentityVerificationOTC from "./emails/IdentityVerificationOTC";

const optsHTML = {
	pretty: false,
	plainText: false,
};

const optsTXT = {
	pretty: false,
	plainText: true,
};

const propsEvent = {
	title: "{{ .Title }}",
	displayName: "{{ .DisplayName }}",
	remoteIP: "{{ .RemoteIP }}",
	detailsKey: "{{ $key }}",
	detailsValue: "{{ index $.Details $key }}",
	detailsPrefix: "{{- $keys := sortAlpha (keys .Details) }}{{- range $key := $keys }}",
	detailsSuffix: "{{ end }}",
};

fs.writeFileSync('../embed/notification/Event.html', render(<Event {...propsEvent} />, optsHTML));
fs.writeFileSync('../embed/notification/Event.txt', render(<Event {...propsEvent} />, optsTXT));

const propsJWT = {
	title: "{{ .Title }}",
	displayName: "{{ .DisplayName }}",
	remoteIP: "{{ .RemoteIP }}",
	link: "{{ .LinkURL }}",
	linkText: "{{ .LinkText }}",
	revocationLinkURL: "{{ .RevocationLinkURL }}",
	revocationLinkText: "{{ .RevocationLinkText }}",
};

fs.writeFileSync('../embed/notification/IdentityVerificationJWT.html', render(<IdentityVerificationJWT {...propsJWT} />, optsHTML));
fs.writeFileSync('../embed/notification/IdentityVerificationJWT.txt', render(<IdentityVerificationJWT {...propsJWT} />, optsTXT));

const propsOTC = {
	title: "{{ .Title }}",
	displayName: "{{ .DisplayName }}",
	remoteIP: "{{ .RemoteIP }}",
	oneTimeCode: "{{ .OneTimeCode }}",
	revocationLinkURL: "{{ .RevocationLinkURL }}",
	revocationLinkText: "{{ .RevocationLinkText }}",
};

fs.writeFileSync('../embed/notification/IdentityVerificationOTC.html', render(<IdentityVerificationOTC {...propsOTC} />, optsHTML));
fs.writeFileSync('../embed/notification/IdentityVerificationOTC.txt', render(<IdentityVerificationOTC {...propsOTC} />, optsTXT));
