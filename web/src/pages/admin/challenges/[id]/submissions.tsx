import { useChallengeApi } from "@/api/challenge";
import { useGameApi } from "@/api/game";
import { useSubmissionApi } from "@/api/submission";
import withChallengeEdit from "@/components/layouts/admin/withChallengeEdit";
import withGameEdit from "@/components/layouts/admin/withGameEdit";
import MDIcon from "@/components/ui/MDIcon";
import { Challenge } from "@/types/challenge";
import { Game } from "@/types/game";
import { Submission } from "@/types/submission";
import {
	Divider,
	Group,
	Stack,
	ThemeIcon,
	Text,
	Table,
	Pagination,
	Flex,
	Badge,
	Avatar,
	ActionIcon,
	Tooltip,
	LoadingOverlay,
} from "@mantine/core";
import dayjs from "dayjs";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";

function Page() {
	const { id } = useParams<{ id: string }>();
	const submissionApi = useSubmissionApi();
	const challengeApi = useChallengeApi();

	const [challenge, setChallenge] = useState<Challenge>();
	const [submissions, setSubmissions] = useState<Array<Submission>>([]);

	const [total, setTotal] = useState<number>(0);
	const [rowsPerPage, setRowsPerPage] = useState<number>(10);
	const [page, setPage] = useState<number>(1);

	const [loading, setLoading] = useState<boolean>(false);

	const statusMap = new Map<number, { color: string; label: string }>([
		[
			0,
			{
				color: "gray",
				label: "Unjudged",
			},
		],
		[
			1,
			{
				color: "red",
				label: "Wrong",
			},
		],
		[
			2,
			{
				color: "green",
				label: "Accpet",
			},
		],
		[
			3,
			{
				color: "orange",
				label: "Cheat",
			},
		],
		[
			4,
			{
				color: "blue",
				label: "Invalid",
			},
		],
	]);

	function getChallenge() {
		challengeApi
			.getChallenges({
				id: Number(id),
			})
			.then((res) => {
				const r = res.data;
				setChallenge(r.data[0]);
			});
	}

	function getSubmissions() {
		setLoading(true);
		submissionApi
			.getSubmissions({
				challenge_id: Number(id),
				page: page,
				size: rowsPerPage,
				is_detailed: true,
			})
			.then((res) => {
				const r = res.data;
				setSubmissions(r.data);
				setTotal(r.total);
			})
			.finally(() => {
				setLoading(false);
			});
	}

	useEffect(() => {
		if (challenge) {
			getSubmissions();
		}
	}, [challenge, page, rowsPerPage]);

	useEffect(() => {
		getChallenge();
	}, []);

	useEffect(() => {
		document.title = `提交记录 - ${challenge?.title}`;
	}, [challenge]);

	return (
		<>
			<Stack m={36}>
				<Stack gap={10}>
					<Group>
						<ThemeIcon variant="transparent">
							<MDIcon>verified</MDIcon>
						</ThemeIcon>
						<Text fw={700} size="xl">
							提交记录
						</Text>
					</Group>
					<Divider />
				</Stack>
				<Stack mx={20} mih={"calc(100vh - 360px)"} pos={"relative"}>
					<LoadingOverlay visible={loading} />
					<Table stickyHeader horizontalSpacing={"md"} striped>
						<Table.Thead>
							<Table.Tr
								sx={{
									lineHeight: 3,
								}}
							>
								<Table.Th />
								<Table.Th>Flag</Table.Th>
								<Table.Th>相关比赛</Table.Th>
								<Table.Th>相关团队</Table.Th>
								<Table.Th>提交者</Table.Th>
								<Table.Th>提交时间</Table.Th>
								<Table.Th />
							</Table.Tr>
						</Table.Thead>
						<Table.Tbody>
							{submissions?.map((submission) => (
								<Table.Tr key={submission?.id}>
									<Table.Td>
										<Badge
											color={
												statusMap?.get(
													Number(submission?.status)
												)?.color
											}
										>
											{
												statusMap?.get(
													Number(submission?.status)
												)?.label
											}
										</Badge>
									</Table.Td>
									<Table.Td
										maw={200}
										sx={{
											overflow: "hidden",
											textOverflow: "ellipsis",
											whiteSpace: "nowrap",
										}}
									>
										{submission?.flag}
									</Table.Td>
									<Table.Td>
										{submission?.game?.title || "练习场"}
									</Table.Td>
									<Table.Td>
										{submission?.team?.name ? (
											<Group gap={15}>
												<Avatar
													color="brand"
													src={`${import.meta.env.VITE_BASE_API}/media/teams/${submission?.team?.id}/${submission?.team?.avatar?.name}`}
													radius="xl"
												>
													<MDIcon>people</MDIcon>
												</Avatar>
												{submission?.team?.name}
											</Group>
										) : (
											"无"
										)}
									</Table.Td>
									<Table.Td>
										<Group gap={15}>
											<Avatar
												color="brand"
												src={`${import.meta.env.VITE_BASE_API}/media/users/${submission?.user?.id}/${submission?.user?.avatar?.name}`}
												radius="xl"
											>
												<MDIcon>person</MDIcon>
											</Avatar>
											{submission?.user?.nickname}
										</Group>
									</Table.Td>
									<Table.Td>
										<Badge>
											{dayjs(
												Number(submission?.created_at)
											).format("YYYY/MM/DD HH:mm:ss")}
										</Badge>
									</Table.Td>
									<Table.Td>
										<Group>
											<Tooltip
												withArrow
												label="删除提交记录"
											>
												<ActionIcon
													variant="transparent"
													color="red"
												>
													<MDIcon>delete</MDIcon>
												</ActionIcon>
											</Tooltip>
										</Group>
									</Table.Td>
								</Table.Tr>
							))}
						</Table.Tbody>
					</Table>
				</Stack>
				<Flex justify={"center"}>
					<Pagination
						withEdges
						total={Math.ceil(total / rowsPerPage)}
						value={page}
						onChange={setPage}
					/>
				</Flex>
			</Stack>
		</>
	);
}

export default withChallengeEdit(Page);