import * as React from 'react';

import {
    Body,
    Container,
    Head,
    Heading,
    Hr,
    Html,
    Preview,
    Section,
    Text,
    Tailwind,
    Button,
    Link,
} from '@react-email/components';

import { Brand } from "../components/Brand";

interface Props {
    title?: string;
    displayName?: string;
    domain?: string;
    remoteIP?: string;
    link?: string;
    linkText?: string;
    revocationLinkURL?: string;
    revocationLinkText?: string;
    hidePreview?: boolean;
}

export const IdentityVerificationJWT = ({
    title,
    displayName,
    domain,
    remoteIP,
    link,
    linkText,
    revocationLinkURL,
    revocationLinkText,
	hidePreview,
}: Props) => {
    return (
        <Html lang="en" dir="ltr">
            <Tailwind>
            <Head />
			{!hidePreview ? (
				<Preview>{title ? title : 'Confirm an action'}</Preview>
			) : null}
                <Body className="bg-white my-auto mx-auto font-sans px-2">
                    <Container className="border border-solid border-[#eaeaea] rounded my-[40px] mx-auto p-[20px] max-w-[465px]">
                        <Heading className="text-black text-[24px] font-normal text-center p-0 my-[30px] mx-0">
                            A <strong>one-time link</strong> has been generated
                            to complete a requested action
                        </Heading>
                        <Text className="text-black text-[14px] leading-[24px]">
                            Hi {displayName},
                        </Text>
                        <Text className="text-black text-[14px] leading-[24px]">
                            We would like to confirm a{' '}
                            <strong>requested action </strong>related to the{' '}
                            <strong>security of your account</strong> at{' '}
                            <i>{domain}</i>
						</Text>
						<Text className="text-black text-[14px] leading-[24px] text-center">
							<strong>
										Do not share this notification or the
										content of this notification with anyone.
							</strong>
						</Text>
                        <Hr className="border border-solid border-[#eaeaea] my-[12px] mx-0 w-full" />
                        <Section className="text-center">
                            <Text className="text-black text-[14px] leading-[24px]">
                                If you made this request, click the validation
                                link below.
                            </Text>
                        </Section>
                        <Section className="text-center">
                            <Button
                                id="link"
                                href={link}
                                className="bg-[#1976d2] rounded text-white text-[12px] font-semibold no-underline text-center px-5 py-3"
                            >
                                {linkText}
                            </Button>
                        </Section>
                        <Text className="text-black text-[14px] leading-[24px] text-center">
                            Alternatively, copy and paste this URL into your
                            browser:
                        </Text>
                        <Section className="text-center">
                            <Link
                                href={link}
								className="text-blue-600 hover:text-blue-800 visited:text-purple-600 text-[12px] no-underline"
								style={{ wordBreak: 'break-all' }}
                            >
                                {link}
                            </Link>
                        </Section>
						<Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
                        <Text>
                            If you did NOT initiate this request, your
                            credentials may have been compromised and you
                            should:
                        </Text>
                        <Section className="text-black text-[14px] leading-[22px]">
                            <ol>
                                <li>
                                    Revoke the validation link using the
                                    provided links below
                                </li>
                                <li>
                                    Reset your password or other login
                                    credentials
                                </li>
                                <li>Contact an Administrator</li>
                            </ol>
                        </Section>
                        <Section className="text-center">
                            <Button
                                id="link-revoke"
                                href={revocationLinkURL}
                                className="bg-[#f50057] rounded text-white text-[12px] font-semibold no-underline text-center px-5 py-3"
                            >
                                {revocationLinkText}
                            </Button>
                        </Section>
                        <Text className="text-black text-[14px] leading-[24px]">
                            To revoke the code click the above button or
                            alternatively copy and paste this URL into your
                            browser:
                        </Text>
                        <Text className="text-black text-[12px] leading-[24px] text-center">
                            <Link
                                href={revocationLinkURL}
								className="text-blue-600 hover:text-blue-800 visited:text-purple-600 text-[12px] no-underline"
								style={{ wordBreak: 'break-all' }}
                            >
                                {revocationLinkURL}
                            </Link>
                        </Text>
                        <Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
                        <Text className="text-[#666666] text-[12px] leading-[24px] text-center">
                            This email was intended for{' '}
                            <span className="text-black">{displayName}</span>.
                            This event was generated due to an action from{' '}
                            <span className="text-black">{remoteIP}</span>. If
                            you do not believe that your actions could have
                            triggered this event or if you are concerned about
                            your account's safety, please follow the explicit
                            directions in this notification.
                        </Text>
                    </Container>
                    <Brand />
                </Body>
            </Tailwind>
        </Html>
    );
};

IdentityVerificationJWT.PreviewProps = {
    title: 'Reset your password',
    displayName: 'John Doe',
    domain: 'example.com',
    link: 'https://auth.example.com/reset-password/step2?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJBdXRoZWxpYSIsImV4cCI6MTcyNzY0MTYwNywiaWF0IjoxNzI3NjM4MDA3LCJqdGkiOiI2NTAwY2NhMy03ZDAzLTQ3NjEtOGFiOC0wN2IwNDcxMTgwMzkiLCJhY3Rpb24iOiJSZXNldFBhc3N3b3JkIiwidXNlcm5hbWUiOiJqb2huIn0.TZjqSAun9uNTIu8680cdGAxuiMC18xiAa80LslTMhqU',
    linkText: 'Validate',
    revocationLinkURL: 'https://auth.example.com/revoke/reset-password?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJBdXRoZWxpYSIsImV4cCI6MTcyNzY0MTYwNywiaWF0IjoxNzI3NjM4MDA3LCJqdGkiOiI2NTAwY2NhMy03ZDAzLTQ3NjEtOGFiOC0wN2IwNDcxMTgwMzkiLCJhY3Rpb24iOiJSZXNldFBhc3N3b3JkIiwidXNlcm5hbWUiOiJqb2huIn0.TZjqSAun9uNTIu8680cdGAxuiMC18xiAa80LslTMhqU',
    revocationLinkText: 'Revoke',
    remoteIP: '127.0.0.1',
	hidePreview: false,
} as Props;

export default IdentityVerificationJWT;
