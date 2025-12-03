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
    oneTimeCode?: string;
    revocationLinkURL?: string;
    revocationLinkText?: string;
	hidePreview?: boolean;
}

export const IdentityVerificationOTC = ({
    title,
    displayName,
    domain,
    remoteIP,
    oneTimeCode,
    revocationLinkURL,
    revocationLinkText,
	hidePreview,
}: Props) => {
    return (
        <Html lang="en" dir="ltr">
            <Tailwind>
            <Head />
			{!hidePreview ? (
				<Preview>
					A one-time code has been generated for session elevation
				</Preview>
			) : null}
                <Body className="bg-white my-auto mx-auto font-sans px-2">
                    <Container className="border border-solid border-[#eaeaea] rounded my-[40px] mx-auto p-[20px] max-w-[465px]">
                        <Heading className="text-black text-[24px] font-normal text-center p-0 my-[30px] mx-0">
                            A <strong>one-time code</strong> has been generated
                            to complete a requested action
                        </Heading>
                        <Text className="text-black text-[14px] leading-[24px]">
                            Hi {displayName},
                        </Text>
                        <Text className="text-black text-[14px] leading-[24px]">
                            This notification has been sent to you in order to
                            verify your identity to{' '}
                            <strong>change security details</strong> for your
                            account at <i>{domain}</i>.{' '}
                        </Text>
                        <Text className="text-black text-[14px] leading-[24px] text-center">
                            <strong>
                                Do not share this notification or the content of
                                this notification with anyone.
                            </strong>
                        </Text>
                        <Text className="text-black text-[14px] leading-[24px]">
                            {' '}
                            The following <i>one-time code</i> should only be
                            used in the prompt displayed in your browser.
                        </Text>

                        <Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
                        <Section>
                            <Text
                                id="one-time-code"
                                className="text-black text-center tracking-[0.5rem] font-bold text-lg"
                                style={{ marginRight: '-0.5rem !important' }}
                            >
                                {oneTimeCode}
                            </Text>
                        </Section>
                        <Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
                        <Text className="text-black text-[14px] leading-[24px]">
                            If you did not initiate the process your credentials
                            may have been compromised and you should:
                        </Text>
                        <Section className="text-black text-[14px] leading-[22px]">
                            <ol>
                                <li>
                                    Revoke this code using the provided links
                                    below
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
                        <Text className="text-black text-[14px] leading-[24px] text-center">
                            To revoke the code click the above button or
                            alternatively copy and paste this URL into your
                            browser:{' '}
                        </Text>
                        <Text className="text-black text-[12px] leading-[24px] text-center">
                            <Link
                                href={revocationLinkURL}
                                className="text-blue-600 no-underline"
                            >
                                {revocationLinkURL}
                            </Link>
                        </Text>
                        <Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
                        <Text className="text-[#666666] text-[12px] leading-[24px] text-center">
                            This notification was intended for{' '}
                            <span className="text-black">{displayName}</span>.
                            This one-time code was generated due to an action
                            from <span className="text-black">{remoteIP}</span>.
                            If you do not believe that your actions could have
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

IdentityVerificationOTC.PreviewProps = {
    title: 'Confirm your identity',
    displayName: 'John Doe',
    domain: 'example.com',
    oneTimeCode: 'ABC123',
    revocationLinkURL: 'https://auth.example.com',
    revocationLinkText: 'Revoke',
    remoteIP: '127.0.0.1',
} as Props;

export default IdentityVerificationOTC;
