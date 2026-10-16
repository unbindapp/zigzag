import { useHotkey } from "@tanstack/react-hotkeys";
import { ChevronRightIcon } from "lucide-react";
import { useEffect, useId, useRef, useState } from "react";
import { useLocalStorage } from "usehooks-ts";

import { useUploadImage } from "@/components/issues/comment-thread";
import {
  applyPatch,
  draftToCreateInput,
  EMPTY_DRAFT,
  type TIssueDraft,
} from "@/components/issues/issue-draft";
import { IssueProperties } from "@/components/issues/issue-properties";
import { type TNewIssueDefaults, useNewIssueStore } from "@/components/issues/new-issue-store";
import {
  useCycles,
  useProjects,
  useTeamStates,
  useWorkspaceUsers,
} from "@/components/issues/picker-sources";
import { PropertyPicker } from "@/components/issues/property-picker";
import { MarkdownEditor } from "@/components/shared/markdown-editor";
import { TeamIcon } from "@/components/shared/team-icon";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogFooter, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Kbd } from "@/components/ui/kbd";
import { Switch } from "@/components/ui/switch";
import { toast } from "@/components/ui/toast";
import { useTeams } from "@/components/workspace/teams-provider";
import { useTeamFromRoute } from "@/components/workspace/use-team-from-route";
import { useCreateIssueMutation } from "@/lib/queries/issues";

const TITLE_MAX_LENGTH = 255;

export function NewIssueDialog() {
  const isOpen = useNewIssueStore((state) => state.isOpen);
  const close = useNewIssueStore((state) => state.close);
  const defaults = useNewIssueStore((state) => state.defaults);
  return (
    <Dialog open={isOpen} onOpenChange={(open) => !open && close()}>
      <DialogContent className="w-2xl max-w-full">
        {isOpen && <NewIssueForm defaults={defaults} onClose={close} />}
      </DialogContent>
    </Dialog>
  );
}

function useInitialTeamId(defaults: TNewIssueDefaults) {
  const { teams, myTeams } = useTeams();
  const routeTeam = useTeamFromRoute();
  const [lastTeamId] = useLocalStorage<string | null>("new-issue-team", null);
  const candidates = [defaults.teamId, routeTeam?.id, lastTeamId, myTeams[0]?.id, teams[0]?.id];
  return candidates.find((id) => id && teams.some((team) => team.id === id));
}

function useResolvedDefaults(teamId: string | undefined, defaults: TNewIssueDefaults) {
  const states = useTeamStates(teamId);
  const { projects } = useProjects(teamId);
  const { cycles } = useCycles(teamId);
  const { teams } = useTeams();
  const team = teams.find((item) => item.id === teamId);
  const provided = Object.fromEntries(
    Object.entries(defaults.draft ?? {}).filter(([, value]) => value !== undefined),
  );
  const draft: TIssueDraft = { ...EMPTY_DRAFT, ...provided };
  if (!draft.state) {
    draft.state =
      states.find((state) => state.id === defaults.stateId) ??
      states.find((state) => state.id === team?.default_state_id);
  }
  if (!draft.project && defaults.projectId) {
    const project = projects.find((item) => item.id === defaults.projectId);
    if (project) {
      draft.project = {
        id: project.id,
        name: project.name,
        slug: project.slug,
        icon: project.icon,
        color: project.color,
      };
    }
  }
  if (!draft.cycle && defaults.cycleId) {
    const cycle = cycles.find((item) => item.id === defaults.cycleId);
    if (cycle) draft.cycle = { id: cycle.id, number: cycle.number, name: cycle.name };
  }
  if (!draft.parent && defaults.parentId) {
    draft.parent = { id: defaults.parentId, identifier: "", title: "Parent issue" };
  }
  return draft;
}

function withoutTeamScoped(defaults: TNewIssueDefaults): TNewIssueDefaults {
  return {
    ...defaults,
    stateId: undefined,
    cycleId: undefined,
    draft: { ...defaults.draft, state: undefined, cycle: undefined, labels: undefined },
  };
}

function NewIssueForm({ defaults, onClose }: { defaults: TNewIssueDefaults; onClose: () => void }) {
  const initialTeamId = useInitialTeamId(defaults);
  const [teamId, setTeamId] = useState<string | undefined>(initialTeamId ?? undefined);
  const [, setLastTeamId] = useLocalStorage<string | null>("new-issue-team", null);
  const resolved = useResolvedDefaults(
    teamId,
    teamId === initialTeamId ? defaults : withoutTeamScoped(defaults),
  );
  const [patch, setPatch] = useState<Partial<TIssueDraft>>({});
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [createMore, setCreateMore] = useState(false);
  const [formKey, setFormKey] = useState(0);
  const titleRef = useRef<HTMLInputElement>(null);
  const submittingRef = useRef(false);
  const createMoreId = useId();
  const { teams, myTeams } = useTeams();
  const { users } = useWorkspaceUsers();
  const uploadImage = useUploadImage();
  const { mutateAsync: createIssue, isPending } = useCreateIssueMutation();
  const draft = applyPatch(resolved, patch);
  const team = teams.find((item) => item.id === teamId);
  const pickableTeams = myTeams.length > 0 ? myTeams : teams;

  useEffect(() => {
    titleRef.current?.focus();
  }, [formKey]);

  const submit = async () => {
    if (!teamId || !title.trim() || submittingRef.current) return;
    submittingRef.current = true;
    const res = await createIssue(draftToCreateInput({ teamId, title, description, draft }))
      .catch(() => null)
      .finally(() => {
        submittingRef.current = false;
      });
    if (!res) return;
    setLastTeamId(teamId);
    toast.add({
      type: "success",
      title: `Created ${res.data.identifier}`,
      description: res.data.title,
    });
    if (!createMore) {
      onClose();
      return;
    }
    setTitle("");
    setDescription("");
    setFormKey((key) => key + 1);
  };

  useHotkey("Mod+Enter", () => void submit(), { ignoreInputs: false, conflictBehavior: "allow" });

  const changeTeam = (nextTeamId: string) => {
    if (nextTeamId === teamId) return;
    setTeamId(nextTeamId);
    setPatch((current) => ({
      ...current,
      state: undefined,
      cycle: null,
      estimate: null,
      milestone: null,
      labels: [],
    }));
  };

  return (
    <form
      className="flex flex-col gap-4"
      onSubmit={(event) => {
        event.preventDefault();
        void submit();
      }}
    >
      <div className="flex items-center gap-2 pr-8">
        <PropertyPicker
          label="Team"
          options={pickableTeams.map((item) => ({
            value: item.id,
            label: item.name,
            keywords: [item.key],
            icon: <TeamIcon teamKey={item.key} icon={item.icon} color={item.color} size="sm" />,
          }))}
          value={teamId ? [teamId] : []}
          onChange={([id]) => changeTeam(id)}
          trigger={
            <Button variant="outline" size="sm" aria-label={`Team: ${team?.name ?? "none"}`}>
              {team && (
                <TeamIcon teamKey={team.key} icon={team.icon} color={team.color} size="sm" />
              )}
              {team?.key ?? "Pick a team"}
            </Button>
          }
        />
        <ChevronRightIcon className="text-muted-foreground size-3.5" />
        <DialogTitle>New issue</DialogTitle>
      </div>
      <Input
        ref={titleRef}
        framed
        aria-label="Issue title"
        placeholder="Issue title"
        maxLength={TITLE_MAX_LENGTH}
        value={title}
        onChange={(event) => setTitle(event.target.value)}
        className="px-0 text-base font-medium"
      />
      <MarkdownEditor
        key={formKey}
        label="Description"
        value={description}
        onChange={setDescription}
        placeholder="Add description…"
        users={users}
        onUploadImage={uploadImage}
      />
      {teamId && (
        <IssueProperties
          teamId={teamId}
          layout="inline"
          value={draft}
          onChange={(next) => setPatch((current) => ({ ...current, ...next }))}
        />
      )}
      <DialogFooter className="items-center border-t pt-4">
        <label
          htmlFor={createMoreId}
          className="text-muted-foreground mr-auto flex items-center gap-2 text-sm"
        >
          <Switch id={createMoreId} checked={createMore} onCheckedChange={setCreateMore} />
          Create more
        </label>
        <Button type="button" variant="outline" size="sm" onClick={onClose}>
          Cancel
        </Button>
        <Button type="submit" size="sm" isPending={isPending} disabled={!teamId || !title.trim()}>
          Create issue
          <Kbd className="bg-primary-foreground/2-10 text-primary-foreground border-transparent">
            ⌘↵
          </Kbd>
        </Button>
      </DialogFooter>
    </form>
  );
}
