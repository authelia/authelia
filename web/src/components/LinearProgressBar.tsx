import { Progress } from "@components/UI/Progress";
import { cn } from "@utils/Styles";

export interface Props {
    value: number;
    height?: number | string;
    className?: string;
}

const LinearProgressBar = function (props: Props) {
    return (
        <Progress
            value={props.value}
            className={cn("mt-2 transition-transform duration-200 linear", props.className)}
            style={{ height: props.height ? props.height : 8 }}
        />
    );
};

export default LinearProgressBar;
