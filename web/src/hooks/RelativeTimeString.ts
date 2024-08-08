import { useEffect, useState } from "react";

function getRelativeTimeString(date: Date): string {
    const now = new Date();
    const secondsSinceUse = (now.getTime() - date.getTime()) / 1000;
    const ONEMINUTE = 60;
    const ONEHOUR = 3600;
    const ONEDAY = 86400;
    const ONEMONTH = 2592000;
    const ONEYEAR = 31536000;

    if (secondsSinceUse < ONEMINUTE) {
        return "just now";
    } else if (secondsSinceUse < ONEHOUR) {
        const minutes = Math.floor(secondsSinceUse / ONEMINUTE);
        return `${minutes} minute${minutes > 1 ? "s" : ""} ago`;
    } else if (secondsSinceUse < ONEDAY) {
        const hours = Math.floor(secondsSinceUse / ONEHOUR);
        return `${hours} hour${hours > 1 ? "s" : ""} ago`;
    } else if (secondsSinceUse < ONEMONTH) {
        const days = Math.floor(secondsSinceUse / ONEDAY);
        return `${days} day${days > 1 ? "s" : ""} ago`;
    } else if (secondsSinceUse < ONEYEAR) {
        const months = Math.floor(secondsSinceUse / ONEMONTH);
        return `${months} month${months > 1 ? "s" : ""} ago`;
    } else if (secondsSinceUse > ONEYEAR) {
        const years = Math.floor(secondsSinceUse / ONEYEAR);
        return years === 1 ? "Over a year ago" : `${years} years ago`;
    } else {
        return "never";
    }
}

function useRelativeTime(date: Date): string {
    const [relativeTime, setRelativeTime] = useState<string>(getRelativeTimeString(date));

    useEffect(() => {
        const intervalId = setInterval(() => {
            setRelativeTime(getRelativeTimeString(date));
        }, 60000); //Every minute
        return () => clearInterval(intervalId);
    }, [date]);
    return relativeTime;
}

export { useRelativeTime, getRelativeTimeString };
