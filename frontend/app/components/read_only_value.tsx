export function ReadOnlyValue({ value }: { value: string }) {
	return (
		<div className="input input-bordered bg-base-100 flex w-full items-center rounded-xl">
			<span className="text-sm wrap-break-word whitespace-pre-wrap opacity-80">{value || '—'}</span>
		</div>
	);
}
