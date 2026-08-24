import { cn } from "@utils/Styles";

export interface Props {
    message: string;
}

const delays = [
    "[animation-delay:0.1s]",
    "[animation-delay:0.2s]",
    "[animation-delay:0.3s]",
    "[animation-delay:0.4s]",
    "[animation-delay:0.5s]",
];

const BaseLoadingPage = function (props: Props) {
    return (
        <div className="grid min-h-screen items-center justify-center">
            <div className="inline-block text-center">
                <div className="p-4">
                    <span className="inline-flex">
                        {delays.map((delay) => (
                            <span
                                key={delay}
                                className={cn(
                                    "mx-0.5 inline-block h-[35px] w-1 rounded-[2px]",
                                    "bg-[var(--custom-loading-bar)] animate-scale-loader",
                                    delay,
                                )}
                            />
                        ))}
                    </span>
                </div>
                <div className="p-4">
                    <p>{props.message}...</p>
                </div>
            </div>
        </div>
    );
};

export default BaseLoadingPage;
