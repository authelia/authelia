import { type MouseEvent, useLayoutEffect, useRef, useState } from "react";

import { Tooltip, TooltipContent, TooltipTrigger } from "@components/UI/Tooltip";

interface TruncatedCellValueProps {
    value: string;
}

function TruncatedCellValue({ value }: Readonly<TruncatedCellValueProps>) {
    const ref = useRef<HTMLSpanElement>(null);
    const [truncated, setTruncated] = useState(false);

    useLayoutEffect(() => {
        const el = ref.current;

        if (!el) {
            return;
        }

        if (el.scrollWidth > el.clientWidth) {
            setTruncated(true);
        } else {
            setTruncated(false);
        }

        if (typeof ResizeObserver === "undefined") {
            return;
        }

        const observer = new ResizeObserver(() => {
            setTruncated(el.scrollWidth > el.clientWidth);
        });

        observer.observe(el);

        return () => observer.disconnect();
    }, [value]);

    const handleTooltipDoubleClick = (event: MouseEvent<HTMLDivElement>) => {
        event.preventDefault();
        event.stopPropagation();
    };

    return (
        <Tooltip>
            <TooltipTrigger
                disabled={!truncated}
                render={
                    <span className="inline-block max-w-full truncate select-text" ref={ref}>
                        {value}
                    </span>
                }
            />
            {truncated ? (
                <TooltipContent
                    className="pointer-events-auto w-auto max-w-sm text-sm break-words select-text"
                    onDoubleClick={handleTooltipDoubleClick}
                    sideOffset={4}
                >
                    {value}
                </TooltipContent>
            ) : null}
        </Tooltip>
    );
}

export { TruncatedCellValue };
