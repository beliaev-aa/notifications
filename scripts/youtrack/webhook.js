const entities = require('@jetbrains/youtrack-scripting-api/entities');
const http = require('@jetbrains/youtrack-scripting-api/http');

const WEBHOOK_URL = 'http://host.docker.internal:3000/log-post-request';

const MENTION_REGEX_FULL = /@\{([^,]+),([^,]+),([^,]+),([^}]+)\}/g;
const MENTION_REGEX_SIMPLE = /@([a-zA-Z0-9._-]+)/g;

/** Создаёт безопасный объект пользователя */
const createUser = user => user ? {
    fullName: user.fullName || null,
    login: user.login || null,
    email: user.email || null
} : null;

/** Формирует URL задачи */
const buildIssueUrl = (ctx, issue) => {
    if (ctx.permalink) return ctx.permalink.startsWith('http') ? ctx.permalink : '/' + ctx.permalink;
    const base = ctx.youTrackBaseUrl || 'http://localhost:8080';
    return issue.idReadable ? `${base}/issue/${issue.idReadable}` : issue.id ? `${base}/issue/${issue.id}` : null;
};

/** Извлекает упомянутых пользователей из текста */
const extractMentions = (ctx, text) => {
    if (!text) return [];
    const users = [], seen = new Set();

    MENTION_REGEX_FULL.lastIndex = 0;
    let m;
    while ((m = MENTION_REGEX_FULL.exec(text)) !== null) {
        if (!seen.has(m[2])) {
            seen.add(m[2]);
            users.push({fullName: m[3], login: m[2], email: m[4]});
        }
    }

    MENTION_REGEX_SIMPLE.lastIndex = 0;
    while ((m = MENTION_REGEX_SIMPLE.exec(text)) !== null) {
        const login = m[1];
        if (seen.has(login)) continue;
        seen.add(login);
        let user = null;
        try {
            user = entities.User.findByLogin(ctx, login);
        } catch {
        }
        if (user) users.push(createUser(user));
    }
    return users;
};

/** Формирует базовый payload */
const basePayload = (ctx, issue, url) => ({
    author: createUser(ctx.currentUser),
    issueUrl: url,
    projectName: issue.project.shortName
});

/** Конфигурация событий: условие + генератор payload */
const EVENTS = [
    {
        check: (ctx, issue) => issue.fields.State.isChanged && (!issue.oldValue('State') || issue.oldValue('State').name !== issue.fields.State.name),
        payload: (ctx, issue, url) => {
            const oldS = issue.oldValue('State'), newS = issue.fields.State;
            return {
                ...basePayload(ctx, issue, url), event: 'status_changed',
                data: {
                    oldStatus: oldS ? {name: oldS.name, presentation: oldS.presentation} : null,
                    newStatus: {name: newS.name, presentation: newS.presentation}
                }
            };
        }
    },
    {
        check: (ctx, issue) => issue.comments.added.isNotEmpty() && issue.comments.added.last()?.text,
        payload: (ctx, issue, url) => {
            const c = issue.comments.added.last();
            const commentUrl = c.url || (url && c.id ? `${url}#comment=${c.id}` : null);
            return {
                ...basePayload(ctx, issue, url), event: 'comment_added',
                data: {timestamp: c.created, text: c.text, mentionedUsers: extractMentions(ctx, c.text), commentUrl}
            };
        }
    },
    {
        check: (ctx, issue) => issue.fields.Assignee && issue.fields.Assignee.isChanged && (issue.oldValue('Assignee')?.fullName !== issue.fields.Assignee.fullName),
        payload: (ctx, issue, url) => ({
            ...basePayload(ctx, issue, url), event: 'assignee_changed',
            data: {oldAssignee: createUser(issue.oldValue('Assignee')), newAssignee: createUser(issue.fields.Assignee)}
        })
    }
];

/** Отправка payload в вебхук */
const sendWebhook = (payload) => {
    try {
        const conn = new http.Connection(WEBHOOK_URL, null, 2000);
        conn.addHeader('Content-Type', 'application/json');
        const resp = conn.postSync('', null, JSON.stringify(payload));
        if (!resp.isSuccess) console.warn(`Webhook failed: ${resp.status}, payload: ${JSON.stringify(payload)}`);
    } catch (err) {
        console.error('Ошибка отправки вебхука:', err, JSON.stringify(payload));
    }
};

exports.rule = entities.Issue.onChange({
    title: 'Send Task Notifications to Webhook',
    action: ctx => {
        const issue = ctx.issue;
        const url = buildIssueUrl(ctx, issue);
        for (const e of EVENTS) {
            if (e.check(ctx, issue)) {
                sendWebhook(e.payload(ctx, issue, url));
                break;
            }
        }
    },
    requirements: {Assignee: {type: entities.User.fieldType}, State: {type: entities.EnumField.fieldType}}
});
