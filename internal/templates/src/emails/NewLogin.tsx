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
	displayName?: string;
	remoteIP?: string;
	deviceInfo?: string;
	rawUserAgent?: string;
	hidePreview?: boolean;
}

export const NewLogin = ({
							 title,
							 date,
							 domain,
							 displayName,
							 remoteIP,
							 deviceInfo,
							 rawUserAgent,
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
							Your account at <i>{domain}</i> was just logged into from a new IP address.
						</Text>
						<Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
						<Section className="m-2">
							<Text><strong>Date:</strong> {date}</Text>
							<Text><strong>IP Address:</strong> {remoteIP}</Text>
							<Text><strong>Device:</strong> {deviceInfo}</Text>
							{rawUserAgent && (
								<Text className="text-[#666666] text-[10px] leading-[14px] break-words">
									<strong>User Agent:</strong> {rawUserAgent}
								</Text>
							)}
						</Section>
						<Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
						<Text className="text-[#666666] text-[12px] leading-[24px] text-center">
							This notification was intended for <span className="text-black">{displayName}</span>. This
							event notification was generated due to an action from <span className="text-black">{remoteIP}</span>.
							If you do not recognize this device or IP address, please change your password immediately and contact an
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
	deviceInfo: "Chrome 121.0.0.0 on Windows 10 (Computer)",
	rawUserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"
} as Props;

export default NewLogin;
