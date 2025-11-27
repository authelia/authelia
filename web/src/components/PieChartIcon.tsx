export interface Props {
    maxProgress?: number;
    progress: number;

    width?: number;
    height?: number;

    color?: string;
    backgroundColor?: string;
}

const PieChartIcon = function (props: Props) {
    const maxProgress = props.maxProgress ? props.maxProgress : 100;
    const width = props.width ? props.width : 20;
    const height = props.height ? props.height : 20;

    const color = props.color ? props.color : "black";
    const backgroundColor = props.backgroundColor ? props.backgroundColor : "white";

    return (
        <svg height={`${width}`} width={`${height}`} viewBox="0 0 26 26">
            <circle r="12" cx="13" cy="13" fill="none" stroke={backgroundColor} strokeWidth="2" />
            <circle r="9" cx="13" cy="13" fill={backgroundColor} stroke="transparent" />
            <circle
                r="5"
                cx="13"
                cy="13"
                fill="none"
                stroke={color}
                strokeWidth="10"
                strokeDasharray={`${props.progress} ${maxProgress}`}
                transform="rotate(-90) translate(-26)"
            />
        </svg>
    );
};

export default PieChartIcon;
