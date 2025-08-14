import type { ErrorInfo, ReactElement } from 'react';
import { Component } from 'react';

interface MdErrProps {
	source: string; // <-- raw markdown
	children: ReactElement;
}
interface MdErrState {
	hasError: boolean;
}

export class MdErrorBoundary extends Component<MdErrProps, MdErrState> {
	public state: MdErrState = { hasError: false };

	public componentDidCatch(err: Error, info: ErrorInfo) {
		console.error('Markdown render error', err, info);
		this.setState({ hasError: true });
	}

	public render() {
		if (this.state.hasError) {
			/* You see the warning + the untouched markdown */
			return (
				<pre className="bg-base-200 text-error p-2 rounded overflow-x-auto">
					{`⚠️ Failed to render message. Showing raw content below:

${this.props.source}`}
				</pre>
			);
		}
		return this.props.children;
	}
}
