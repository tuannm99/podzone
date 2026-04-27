import { PageShell } from '../components/common/PageShell';
import { Button, Card } from '../components/common/Primitives';
import { SectionLead } from '../components/common/SectionLead';

export function NotFoundPage(props: {
  navigate: (to: string) => void;
  path: string;
}) {
  return (
    <PageShell>
      <Card class="space-y-6">
        <SectionLead
          eyebrow="Route not found"
          title="That page does not exist."
          copy={`The current path "${props.path}" is not wired into the Solid app.`}
        />
        <div class="flex justify-start">
          <Button pill onClick={() => props.navigate('/admin')}>
            Go back to admin
          </Button>
        </div>
      </Card>
    </PageShell>
  );
}
