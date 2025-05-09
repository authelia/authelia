import * as React from "react";

import {
	Body,
	Container,
	Head,
	Hr,
	Html,
	Preview,
	Section,
	Text,
	Tailwind
} from "@react-email/components";

import { Brand } from "../components/Brand";

export interface Props {
	title?: string;
	date?: string;
	domain?: string;
	userAgent?: string;
	displayName?: string;
	remoteIP?: string;
	hidePreview?: boolean;
}

export const NewLogin = ({
						  title,
						  date,
						  domain,
						  displayName,
						  remoteIP,
						  userAgent,
	                      hidePreview,
					  }: Props) => {
	return (
		<Html lang="en" dir="ltr">
			<Head />
			{!hidePreview ? (
				<Preview>Login from New IP</Preview>
			) : null}
			<Tailwind>
				<Body className="bg-white my-auto mx-auto font-sans px-2">
					<Container className="border border-solid border-[#eaeaea] rounded my-[40px] mx-auto p-[20px] max-w-[465px]">
						{title ? <Text className="text-black text-[24px] font-normal text-center p-0 my-[30px] mx-0">{title}</Text> : null}
						<Text className="text-black text-[14px] leading-[24px]">
							Hi {displayName},
						</Text>
						<Text className="text-black text-[14px] leading-[24px]">
							Your account at <i>{domain}</i> was just logged into from a new ip.
						</Text>
						<Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
						<Section className="m-2">
							<Text><strong>Date:</strong> {date}</Text>
							<Text><strong>IP:</strong> {remoteIP}</Text>
							<Text><strong>User Agent:</strong> {userAgent}</Text>
						</Section>
						<Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
						<Text className="text-[#666666] text-[12px] leading-[24px] text-center">
							This notification was intended for <span className="text-black">{displayName}</span>. This
							event notification was generated due to an action from <span className="text-black">{remoteIP}</span>.
							If you do not believe that your actions could have triggered this event or if you are
							concerned about your account's safety, please change your password and reach out to an
							administrator.
						</Text>
					</Container>
					<Brand />
				</Body>
			</Tailwind>
		</Html>
	);
};

NewLogin.PreviewProps = {
	displayName: "John Doe",
	domain: "example.com",
	date: "Friday, May 9, 2025 at 00:45:10 AM +08:00",
	title: "Login From New IP",
	remoteIP: "127.0.0.1",
	userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
} as Props;
/**
 * Its probably worth doing some sort of user agent parsing and spitting out a "device os/platform" instead of the raw user agent.
 */

export default NewLogin;
