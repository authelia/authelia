import { CircleCheck } from "lucide-react";

const glyph =
    "[&>path]:[transform-box:view-box] [&>path]:[transform-origin:12px_12px] [&>path]:scale-[1.4] " +
    "text-[oklch(from_var(--success)_0.52_c_h)]";

const SuccessIcon = function () {
    return <CircleCheck className={`success-icon size-17.5 ${glyph}`} />;
};

export default SuccessIcon;
