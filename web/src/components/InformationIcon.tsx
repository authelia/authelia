import { Info } from "lucide-react";

export interface Props {}

const glyph =
    "[&>path]:[transform-box:view-box] [&>path]:[transform-origin:12px_12px] [&>path]:scale-[1.4] " +
    "text-[oklch(from_var(--info)_0.55_c_h)]";

const InformationIcon = function () {
    return <Info className={`information-icon size-17.5 ${glyph}`} />;
};

export default InformationIcon;
