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
    bodyEvent?: string;
    bodyPrefix?: string;
    bodySuffix?: string;
	displayName?: string;
	remoteIP?: string;
	detailsKey?: string;
	detailsValue?: string;
	detailsPrefix?: string;
	detailsSuffix?: string;
	hidePreview?: boolean;
}

export const Event = ({
						  title,
                          bodyEvent,
                          bodyPrefix,
                          bodySuffix,
						  displayName,
						  remoteIP,
						  detailsKey,
						  detailsValue,
						  detailsPrefix,
						  detailsSuffix,
	                      hidePreview,
					  }: Props) => {
	return (
		<Html lang="en" dir="ltr">
            <Tailwind>
            <Head />
			{!hidePreview ? (
				<Preview>An important event has occurred with your account</Preview>
			) : null}
				<Body className="bg-white my-auto mx-auto font-sans px-2">
					<Container className="border border-solid border-[#eaeaea] rounded my-[40px] mx-auto p-[20px] max-w-[465px]">
						{title ? <Text className="text-black text-[24px] font-normal text-center p-0 my-[30px] mx-0">{title}</Text> : null}
						<Text className="text-black text-[14px] leading-[24px]">
							Hi {displayName},
						</Text>
						<Text className="text-black text-[14px] leading-[24px]">
							This notification has been sent to you in order to notify you that {bodyPrefix} <strong><i>{bodyEvent}</i></strong> {bodySuffix}
						</Text>
						<Hr className="border border-solid border-[#eaeaea] my-[26px] mx-0 w-full" />
						<Text>Event Details:</Text>
						<Section className="m-2">
							{detailsPrefix}
							<Text><strong>{detailsKey}:</strong> {detailsValue}</Text>
							{detailsSuffix}
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

Event.PreviewProps = {
	displayName: "John Doe",
	detailsKey: "Example Detail",
	detailsValue: "Example Value",
	title: "Second Factor Method Added",
    bodyEvent: "Second Factor Method",
    bodyPrefix: "a",
    bodySuffix: "was added to your account.",
	remoteIP: "127.0.0.1",
} as Props;

export default Event;
