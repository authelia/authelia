import AppleStore from "@assets/images/applestore-badge.svg";
import GooglePlay from "@assets/images/googleplay-badge.svg";
import { cn } from "@utils/Styles";

export interface Props {
    iconSize: number;
    googlePlayLink: string;
    appleStoreLink: string;

    targetBlank?: boolean;
    className?: string;
}

const AppStoreBadges = function (props: Props) {
    const target = props.targetBlank ? "_blank" : undefined;

    return (
        <div className={cn("flex items-center justify-center gap-2", props.className)}>
            <a href={props.googlePlayLink} target={target} className="hover:underline">
                <img src={GooglePlay} alt="google play" width={props.iconSize} />
            </a>
            <a href={props.appleStoreLink} target={target} className="hover:underline">
                <img src={AppleStore} alt="apple store" width={props.iconSize} />
            </a>
        </div>
    );
};

export default AppStoreBadges;
