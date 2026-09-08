import { CircleX } from "lucide-react";

export interface Props {}

const glyph =
    "[&>path]:[transform-box:view-box] [&>path]:[transform-origin:12px_12px] [&>path]:scale-[1.4] " +
    "text-[oklch(from_var(--destructive)_0.55_c_h)]";

const FailureIcon = function () {
    return <CircleX className={`failure-icon size-17.5 ${glyph}`} />;
};

export default FailureIcon;
