import { useEffect, useState } from "react";

export function useIntermittentClass(
    classname: string,
    activeMilliseconds: number,
    inactiveMillisecond: number,
    startMillisecond?: number,
) {
    const [currentClass, setCurrentClass] = useState("");
    const [firstTime, setFirstTime] = useState(true);

    useEffect(() => {
        let timeout: NodeJS.Timeout;

        if (firstTime) {
            if (startMillisecond && startMillisecond > 0) {
                timeout = setTimeout(() => {
                    setCurrentClass(classname);
                    setFirstTime(false);
                }, startMillisecond);
            } else {
                timeout = setTimeout(() => {
                    setCurrentClass(classname);
                    setFirstTime(false);
                }, 0);
            }
        } else {
            if (currentClass === "") {
                timeout = setTimeout(() => setCurrentClass(classname), inactiveMillisecond);
            } else {
                timeout = setTimeout(() => setCurrentClass(""), activeMilliseconds);
            }
        }
        return () => clearTimeout(timeout);
    }, [currentClass, classname, activeMilliseconds, inactiveMillisecond, startMillisecond, firstTime]);

    return currentClass;
}
