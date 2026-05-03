import { useEffect, useRef, useState } from "react";

export function useIntermittentClass(
    classname: string,
    activeMilliseconds: number,
    inactiveMillisecond: number,
    startMillisecond?: number,
) {
    const [currentClass, setCurrentClass] = useState("");
    const firstTimeRef = useRef(true);

    useEffect(() => {
        let timeout: NodeJS.Timeout;

        if (currentClass === "") {
            const delay = firstTimeRef.current ? (startMillisecond ?? 0) : inactiveMillisecond;
            timeout = setTimeout(() => {
                setCurrentClass(classname);
                firstTimeRef.current = false;
            }, delay);
        } else {
            timeout = setTimeout(() => setCurrentClass(""), activeMilliseconds);
        }

        return () => clearTimeout(timeout);
    }, [currentClass, classname, activeMilliseconds, inactiveMillisecond, startMillisecond]);

    return currentClass;
}
