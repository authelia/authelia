/**
 * @module RelativeTimeString
 * @description This module provides utilities for generating and updating relative time strings.
 */

import { useEffect, useState } from "react";

import i18next from "i18next";

/**
 * Time constants in seconds
 * @constant
 */
const ONEMINUTE = 60;
const ONEHOUR = 3600;
const ONEDAY = 86400;
const ONEMONTH = 2592000;
const ONEYEAR = 31536000;

/**
 *
 * @function
 * @param {Date} date - The date used to compare against the current time.
 * @returns {string} A human-readable string representing the time since the date specified. Returned in the current browser locale.
 *
 * @example
 * // Returns "2 hours ago" if the date was 2 hours before the current time
 * const relativeTime = getRelativeTimeString(new Date(Date.now() - 2 * 60 * 60 * 1000));
 */
export function getRelativeTimeString(date: Date): string {
    const now = new Date();
    const secondsSinceUse = (now.getTime() - date.getTime()) / 1000;

    if (secondsSinceUse < ONEMINUTE) {
        return new Intl.RelativeTimeFormat(i18next.languages, { numeric: "auto" }).format(
            0 - Math.floor(secondsSinceUse / ONEMINUTE),
            "seconds",
        );
    } else if (secondsSinceUse < ONEHOUR) {
        return new Intl.RelativeTimeFormat(i18next.languages, { numeric: "auto" }).format(
            0 - Math.floor(secondsSinceUse / ONEMINUTE),
            "minutes",
        );
    } else if (secondsSinceUse < ONEDAY) {
        return new Intl.RelativeTimeFormat(i18next.languages, { numeric: "auto" }).format(
            0 - Math.floor(secondsSinceUse / ONEHOUR),
            "hours",
        );
    } else if (secondsSinceUse < ONEMONTH) {
        return new Intl.RelativeTimeFormat(i18next.languages, { numeric: "auto" }).format(
            0 - Math.floor(secondsSinceUse / ONEDAY),
            "days",
        );
    } else if (secondsSinceUse < ONEYEAR) {
        return new Intl.RelativeTimeFormat(i18next.languages, { numeric: "auto" }).format(
            0 - Math.floor(secondsSinceUse / ONEMONTH),
            "months",
        );
    } else if (secondsSinceUse > ONEYEAR) {
        return new Intl.RelativeTimeFormat(i18next.languages, { numeric: "auto" }).format(
            0 - Math.floor(secondsSinceUse / ONEYEAR),
            "years",
        );
    } else {
        return "never";
    }
}

/**
 *
 * @function
 * @param {Date} date - The date to compare against the current time
 * @returns {string} A human readable string representing the time since the date specified, updated every minute. Returned in the current browser locale.
 *
 * @example
 * function LastLoginTime({ loginDate }){
 *  const relativeTime = useRelativeTime(loginDate);
 *  return <span>Last Login: {relativeTime}</span>;
 * }
 */
export function useRelativeTime(date: Date): string {
    const [relativeTime, setRelativeTime] = useState<string>(getRelativeTimeString(date));

    useEffect(() => {
        const intervalId = setInterval(() => {
            setRelativeTime(getRelativeTimeString(date));
        }, 60000); //Every minute
        return () => clearInterval(intervalId);
    }, [date]);
    return relativeTime;
}
