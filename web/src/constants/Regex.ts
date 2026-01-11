/**
 *  Username, Groups, and Display_Name should be a max of 100 characters.
 *  Usernames are allowed to be either alphanumeric with ,-_ OR a valid email.
 *  Display_Name should be all printable Unicode characters (non-control characters)
 **/
export const REGEX = {
    DISPLAY_NAME: /^[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]{1,100}$/u,
    EMAIL: /^[a-zA-Z0-9+_~!#$%&'*/=?^{|}\-.]+@[a-zA-Z0-9-.]+\.[a-zA-Z0-9-]+$/,
    GROUP: /^[a-zA-Z0-9-_,]{1,100}$/,
    USERNAME: /^[a-zA-Z0-9-_,]{1,100}$/,
};
