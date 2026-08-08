import {type ChangeEvent, forwardRef} from "react";
import ControlStyles from "../control/Control.module.css";
import Styles from "./Input.module.css";
import {Label} from "../label";

type TError = {
    showError: boolean;
    errorMessage: string;
};

type TInputProps = {
    label?: string;
    name?: string;
    type?: 'text' | 'email' | 'password';
    value: string;
    placeholder?: string;
    onChange?: (value: string) => void;
    error?: TError;
    disabled?: boolean,
    loading?: boolean,
}

export const Input = forwardRef<HTMLInputElement, TInputProps>(({
                                                                    label,
                                                                    name,
                                                                    type = 'text',
                                                                    value = "",
                                                                    placeholder,
                                                                    onChange,
                                                                    error,
                                                                    disabled = false,
                                                                    loading = false
                                                                }, ref) => {
    const inputClasses = [
        Styles['input'],
        ControlStyles['text'],
        (disabled || loading) && Styles['input--disabled'],
        error?.showError && Styles['input--error']
    ].filter(Boolean).join(' ');

    const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
        onChange?.(e.target.value);
    };

    return (
        <Label label={label} error={error} disabled={disabled}>
            <input
                className={inputClasses}
                value={value}
                name={name}
                type={type}
                placeholder={placeholder}
                disabled={disabled || loading}
                onChange={handleChange}
                ref={ref}
            />
        </Label>
    )
});

Input.displayName = "Input";
