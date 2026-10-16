import {
  type AnyFieldApi,
  type AnyFormApi,
  createFormHook,
  createFormHookContexts,
  type FormAsyncValidateOrFn,
  type FormOptions,
  type FormValidateOrFn,
  useStore,
} from "@tanstack/react-form";
import type React from "react";
import { type ComponentProps, type ReactNode, useId } from "react";

import { ColorPicker } from "@/components/shared/color-picker";
import { IconPicker } from "@/components/shared/icon-picker";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";

const { fieldContext, formContext, useFieldContext, useFormContext } = createFormHookContexts();

const formDomIds = new WeakMap<AnyFormApi["store"], string>();

function scrollToFirstInvalidField(formDomId: string) {
  requestAnimationFrame(() => {
    const element = document.querySelector<HTMLElement>(
      `[data-form-id="${formDomId}"][data-invalid]`,
    );
    if (!element) return;
    element.scrollIntoView({ behavior: "smooth", block: "center" });
    element
      .querySelector<HTMLElement>("input, textarea, select, button, [tabindex]")
      ?.focus({ preventScroll: true });
  });
}

function errorMessage(error: unknown): string | undefined {
  if (!error) return undefined;
  if (typeof error === "string") return error;
  if (typeof error === "object" && "message" in error && typeof error.message === "string") {
    return error.message;
  }
  return undefined;
}

export function useFieldError(field: AnyFieldApi) {
  const submissionAttempts = useStore(field.form.store, (state) => state.submissionAttempts);
  const isFormSubmitted = submissionAttempts > 0;
  const shown =
    ((field.state.meta.isTouched && field.state.meta.isBlurred) || isFormSubmitted) &&
    field.state.meta.errors.length > 0;
  return {
    error: shown ? errorMessage(field.state.meta.errors[0]) : undefined,
    formDomId: formDomIds.get(field.form.store),
  };
}

type TFieldShellProps = {
  label: string;
  description?: string;
  hideLabel?: boolean;
};

function FieldShell({
  id,
  label,
  description,
  hideLabel,
  layout = "stacked",
  children,
}: TFieldShellProps & { id: string; layout?: "stacked" | "inline"; children: ReactNode }) {
  const field = useFieldContext<unknown>();
  const { error, formDomId } = useFieldError(field);
  const descriptionId = description ? `${id}-description` : undefined;
  const errorId = error ? `${id}-error` : undefined;

  if (layout === "inline") {
    return (
      <div
        data-form-id={formDomId}
        data-invalid={error ? true : undefined}
        className="flex w-full items-start gap-4"
      >
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <label htmlFor={id} className="text-sm font-medium">
            {label}
          </label>
          {description && (
            <p id={descriptionId} className="text-muted-foreground text-sm">
              {description}
            </p>
          )}
          {error && (
            <p id={errorId} role="alert" className="text-destructive text-xs">
              {error}
            </p>
          )}
        </div>
        <div className="flex shrink-0 items-center">{children}</div>
      </div>
    );
  }

  return (
    <div
      data-form-id={formDomId}
      data-invalid={error ? true : undefined}
      className="flex min-w-0 flex-col gap-1.5"
    >
      <label htmlFor={id} className={hideLabel ? "sr-only" : "text-sm font-medium"}>
        {label}
      </label>
      {children}
      {description && (
        <p id={descriptionId} className="text-muted-foreground text-xs">
          {description}
        </p>
      )}
      {error && (
        <p id={errorId} role="alert" className="text-destructive text-xs">
          {error}
        </p>
      )}
    </div>
  );
}

function useFieldIds(description?: string) {
  const id = useId();
  return { id, describedBy: description ? `${id}-description` : undefined };
}

type TTextFieldProps = TFieldShellProps &
  Omit<ComponentProps<"input">, "value" | "onChange" | "onBlur" | "id" | "name"> & {
    transform?: (value: string) => string;
  };

function TextField({ label, description, hideLabel, transform, ...inputProps }: TTextFieldProps) {
  const field = useFieldContext<string>();
  const { id, describedBy } = useFieldIds(description);
  const { error } = useFieldError(field);
  return (
    <FieldShell id={id} label={label} description={description} hideLabel={hideLabel}>
      <Input
        {...inputProps}
        id={id}
        name={field.name}
        value={field.state.value ?? ""}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy}
        onBlur={field.handleBlur}
        onChange={(event) =>
          field.handleChange(transform ? transform(event.target.value) : event.target.value)
        }
      />
    </FieldShell>
  );
}

type TTextareaFieldProps = TFieldShellProps &
  Omit<ComponentProps<"textarea">, "value" | "onChange" | "onBlur" | "id" | "name">;

function TextareaField({ label, description, hideLabel, ...textareaProps }: TTextareaFieldProps) {
  const field = useFieldContext<string>();
  const { id, describedBy } = useFieldIds(description);
  const { error } = useFieldError(field);
  return (
    <FieldShell id={id} label={label} description={description} hideLabel={hideLabel}>
      <Textarea
        {...textareaProps}
        id={id}
        name={field.name}
        value={field.state.value ?? ""}
        aria-invalid={error ? true : undefined}
        aria-describedby={describedBy}
        onBlur={field.handleBlur}
        onChange={(event) => field.handleChange(event.target.value)}
      />
    </FieldShell>
  );
}

export type TSelectFieldItem = {
  value: string;
  label: string;
  icon?: ReactNode;
};

type TSelectFieldProps = TFieldShellProps & {
  items: TSelectFieldItem[];
  placeholder?: string;
  disabled?: boolean;
};

function SelectField({
  label,
  description,
  hideLabel,
  items,
  placeholder = "Select…",
  disabled,
}: TSelectFieldProps) {
  const field = useFieldContext<string>();
  const { id, describedBy } = useFieldIds(description);
  const { error } = useFieldError(field);
  const selected = items.find((item) => item.value === field.state.value);
  return (
    <FieldShell id={id} label={label} description={description} hideLabel={hideLabel}>
      <Select
        items={items}
        value={field.state.value ?? null}
        disabled={disabled}
        onValueChange={(value) => {
          field.handleChange(value ?? "");
          field.handleBlur();
        }}
      >
        <SelectTrigger
          id={id}
          aria-invalid={error ? true : undefined}
          aria-describedby={describedBy}
        >
          <SelectValue placeholder={placeholder}>
            {selected && (
              <>
                {selected.icon}
                <span className="truncate">{selected.label}</span>
              </>
            )}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          {items.map((item) => (
            <SelectItem key={item.value} value={item.value}>
              {item.icon}
              <span className="truncate">{item.label}</span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </FieldShell>
  );
}

type TSwitchFieldProps = TFieldShellProps & { disabled?: boolean };

function SwitchField({ label, description, disabled }: TSwitchFieldProps) {
  const field = useFieldContext<boolean>();
  const { id, describedBy } = useFieldIds(description);
  return (
    <FieldShell id={id} label={label} description={description} layout="inline">
      <Switch
        id={id}
        checked={field.state.value ?? false}
        disabled={disabled}
        aria-describedby={describedBy}
        onCheckedChange={(checked) => {
          field.handleChange(checked);
          field.handleBlur();
        }}
      />
    </FieldShell>
  );
}

type TColorFieldProps = TFieldShellProps & { disabled?: boolean };

function ColorField({ label, description, disabled }: TColorFieldProps) {
  const field = useFieldContext<string>();
  const { id } = useFieldIds(description);
  return (
    <FieldShell id={id} label={label} description={description}>
      <div className="flex">
        <ColorPicker
          value={field.state.value}
          disabled={disabled}
          label={label}
          onChange={(color) => {
            field.handleChange(color);
            field.handleBlur();
          }}
        />
      </div>
    </FieldShell>
  );
}

type TIconFieldProps = TFieldShellProps & { disabled?: boolean; color?: string | null };

function IconField({ label, description, disabled, color }: TIconFieldProps) {
  const field = useFieldContext<string>();
  const { id } = useFieldIds(description);
  return (
    <FieldShell id={id} label={label} description={description}>
      <div className="flex">
        <IconPicker
          value={field.state.value}
          color={color}
          disabled={disabled}
          label={label}
          onChange={(icon) => {
            field.handleChange(icon ?? "");
            field.handleBlur();
          }}
        />
      </div>
    </FieldShell>
  );
}

type TSubmitButtonProps = Omit<ComponentProps<typeof Button>, "type" | "isPending"> & {
  requireDirty?: boolean;
};

function SubmitButton({ requireDirty, disabled, ...props }: TSubmitButtonProps) {
  const form = useFormContext();
  const { isSubmitting, isDirty } = useStore(form.store, (state) => ({
    isSubmitting: state.isSubmitting,
    isDirty: state.isDirty,
  }));
  return (
    <Button
      type="submit"
      isPending={isSubmitting}
      disabled={disabled || (requireDirty && !isDirty)}
      {...props}
    />
  );
}

const {
  useAppForm: useAppFormBase,
  withForm,
  withFieldGroup,
} = createFormHook({
  fieldComponents: {
    TextField,
    TextareaField,
    SelectField,
    SwitchField,
    ColorField,
    IconField,
  },
  formComponents: {
    SubmitButton,
  },
  fieldContext,
  formContext,
});

export function useAppForm<
  TFormData,
  TOnMount extends undefined | FormValidateOrFn<TFormData>,
  TOnChange extends undefined | FormValidateOrFn<TFormData>,
  TOnChangeAsync extends undefined | FormAsyncValidateOrFn<TFormData>,
  TOnBlur extends undefined | FormValidateOrFn<TFormData>,
  TOnBlurAsync extends undefined | FormAsyncValidateOrFn<TFormData>,
  TOnSubmit extends undefined | FormValidateOrFn<TFormData>,
  TOnSubmitAsync extends undefined | FormAsyncValidateOrFn<TFormData>,
  TOnDynamic extends undefined | FormValidateOrFn<TFormData>,
  TOnDynamicAsync extends undefined | FormAsyncValidateOrFn<TFormData>,
  TOnServer extends undefined | FormAsyncValidateOrFn<TFormData>,
  TSubmitMeta,
>(
  options: FormOptions<
    TFormData,
    TOnMount,
    TOnChange,
    TOnChangeAsync,
    TOnBlur,
    TOnBlurAsync,
    TOnSubmit,
    TOnSubmitAsync,
    TOnDynamic,
    TOnDynamicAsync,
    TOnServer,
    TSubmitMeta
  >,
) {
  const formDomId = useId();
  const form = useAppFormBase({
    ...options,
    onSubmitInvalid: (props) => {
      scrollToFirstInvalidField(formDomId);
      options.onSubmitInvalid?.(props);
    },
  });
  formDomIds.set(form.store, formDomId);
  return form;
}

export { withFieldGroup, withForm };

export function submitHandler(form: { handleSubmit: () => Promise<unknown> }) {
  return (event: React.FormEvent) => {
    event.preventDefault();
    event.stopPropagation();
    form.handleSubmit().catch(() => undefined);
  };
}
