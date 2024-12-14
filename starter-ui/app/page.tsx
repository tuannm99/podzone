'use client';

import * as React from 'react';
import { useForm } from 'react-hook-form';

type InputForm = {
  firstName: string;
  lastName: string;
};

export default function App() {
  const {
    register,
    setValue,
    handleSubmit,
    // formState: { errors },
  } = useForm<InputForm>();
  const onSubmit = handleSubmit((data) => console.log(data));

  return (
    <form onSubmit={onSubmit}>
      <label className="label">First Name</label>
      <input className="input" {...register('firstName')} />
      <label className="label">Last Name</label>
      <input className="input" {...register('lastName')} />
      <button
        className="drawer-button"
        type="button"
        onClick={() => {
          setValue('lastName', 'fixed lastname');
        }}
      >
        SetValue
      </button>
    </form>
  );
}
