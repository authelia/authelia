import { ScaleLoader } from "react-spinners";

export interface Props {
    message: string;
}

const BaseLoadingPage = function (props: Props) {
    return (
        <div className="grid min-h-screen items-center justify-center">
            <div className="inline-block text-center">
                <div className="p-4">
                    <ScaleLoader color="var(--custom-loading-bar)" speedMultiplier={1.5} />
                </div>
                <div className="p-4">
                    <p>{props.message}...</p>
                </div>
            </div>
        </div>
    );
};

export default BaseLoadingPage;
