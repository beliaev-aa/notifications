const entities = require('@jetbrains/youtrack-scripting-api/entities');
const http = require('@jetbrains/youtrack-scripting-api/http');

const WEBHOOK_URL = 'http://host.docker.internal:3000/webhook/youtrack';

const MENTION_REGEX_FULL = /@\{([^,]+),([^,]+),([^,]+),([^}]+)\}/g;
const MENTION_REGEX_SIMPLE = /@([a-zA-Z0-9._-]+)/g;

/** Создаёт безопасный объект пользователя */
const createUser = user => user ? {
    fullName: user.fullName || null,
    login: user.login || null,
    email: user.email || null
} : null;

/** Извлекает упомянутых пользователей из текста */
const extractMentions = (ctx, text) => {
    if (!text) return [];
    const users = [], seen = new Set();

    MENTION_REGEX_FULL.lastIndex = 0;
    let m;
    while ((m = MENTION_REGEX_FULL.exec(text)) !== null) {
        if (!seen.has(m[2])) {
            seen.add(m[2]);
            users.push(createUser({
                fullName: m[3],
                login: m[2],
                email: m[4]
            }));
        }
    }

    MENTION_REGEX_SIMPLE.lastIndex = 0;
    while ((m = MENTION_REGEX_SIMPLE.exec(text)) !== null) {
        const login = m[1];
        if (seen.has(login)) continue;
        seen.add(login);
        let user = null;
        try {
            user = entities.User.findByLogin(login);
        } catch (err) {
            // Пользователь не найден или произошла ошибка - просто игнорируем
            user = null;
        }
        // Добавляем пользователя только если он найден в YouTrack
        if (user) {
            users.push(createUser(user));
        }
    }
    return users;
};

/** Формирует массив изменений */
const buildChanges = (ctx, issue) => {
        const changes = [];

        // Изменение статуса
        if (issue.fields.isChanged(ctx.State)) {
            const oldState = issue.oldValue('State');
            const newState = issue.fields.State;

            const oldName = oldState ? (oldState.name || null) : null;
            const newName = newState ? (newState.name || null) : null;

            // Добавляем только если значение реально изменилось
            if (oldName !== newName && oldName !== null) {
                const oldValueObj = oldState ? {
                    name: oldState.name || null,
                    presentation: oldState.presentation || null
                } : null;
                const newValueObj = newState ? {
                    name: newState.name || null,
                    presentation: newState.presentation || null
                } : null;

                changes.push({
                    field: 'State',
                    oldValue: oldValueObj,
                    newValue: newValueObj
                });
            }
        }

        // Изменение приоритета
        if (issue.fields.isChanged(ctx.Priority)) {
            const oldPriority = issue.oldValue('Priority');
            const newPriority = issue.fields.Priority;

            const oldName = oldPriority ? (oldPriority.name || null) : null;
            const newName = newPriority ? (newPriority.name || null) : null;

            // Добавляем только если значение реально изменилось
            if (oldName !== newName && oldName !== null) {
                const oldValueObj = oldPriority ? {
                    name: oldPriority.name || null,
                    presentation: oldPriority.presentation || null
                } : null;
                const newValueObj = newPriority ? {
                    name: newPriority.name || null,
                    presentation: newPriority.presentation || null
                } : null;

                changes.push({
                    field: 'Priority',
                    oldValue: oldValueObj,
                    newValue: newValueObj
                });
            }
        }

        // Изменение исполнителя
        if (issue.fields.isChanged(ctx.Assignee)) {
            const oldAssignee = issue.oldValue('Assignee');
            const newAssignee = issue.fields.Assignee;

            const oldValueObj = oldAssignee ? createUser(oldAssignee) : null;
            const newValueObj = newAssignee ? createUser(newAssignee) : null;

            changes.push({
                field: 'Assignee',
                oldValue: oldValueObj,
                newValue: newValueObj
            });
        }

        // Добавление комментария (только если был добавлен новый комментарий)
        if (issue.comments.added.isNotEmpty()) {
            const lastComment = issue.comments.added.last();
            if (lastComment && lastComment.text) {
                const mentions = extractMentions(ctx, lastComment.text);
                // Формируем объект комментария с упомянутыми пользователями
                const commentChange = {
                    field: 'Comment',
                    oldValue: null,
                    newValue: {
                        text: lastComment.text,
                        mentionedUsers: mentions.length > 0 ? mentions : null
                    }
                };
                changes.push(commentChange);
            }
        }

        return changes;
    }
;

/** Формирует payload для отправки в webhook */
const buildPayload = (ctx, issue) => {
    const changes = buildChanges(ctx, issue);

    // Формируем объект проекта
    const project = {
        name: issue.project.key || null,
        presentation: issue.project.name || null
    };

    // Формируем объект задачи
    const issueObj = {
        idReadable: issue.idReadable || '',
        summary: issue.summary || '',
        url: issue.url
    };

    // Статус задачи (может быть null)
    if (issue.fields.State) {
        issueObj.state = {
            name: issue.fields.State.name || null,
            presentation: issue.fields.State.presentation || null
        };
    } else {
        issueObj.state = {
            name: null,
            presentation: null
        };
    }

    // Приоритет задачи (может быть null)
    if (issue.fields.Priority) {
        issueObj.priority = {
            name: issue.fields.Priority.name || null,
            presentation: issue.fields.Priority.presentation || null
        };
    } else {
        issueObj.priority = {
            name: null,
            presentation: null
        };
    }

    // Исполнитель задачи (может быть null)
    if (issue.fields.Assignee) {
        issueObj.assignee = createUser(issue.fields.Assignee);
    } else {
        issueObj.assignee = null;
    }

    // Автор изменения
    const updater = createUser(ctx.currentUser);

    return {
        project: project,
        issue: issueObj,
        updater: updater,
        changes: changes
    };
};

/** Отправка payload в вебхук */
const sendWebhook = (payload) => {
    try {
        const conn = new http.Connection(WEBHOOK_URL, null, 2000);
        conn.addHeader('Content-Type', 'application/json');
        const resp = conn.postSync('', null, JSON.stringify(payload));
        if (!resp.isSuccess) {
            console.warn(`Webhook failed: ${resp.status}, payload: ${JSON.stringify(payload)}`);
        }
    } catch (err) {
        console.error('Ошибка отправки вебхука:', err, JSON.stringify(payload));
    }
};

exports.rule = entities.Issue.onChange({
    title: 'Send Task Notifications to Webhook',
    action: ctx => {
        const issue = ctx.issue;
        const payload = buildPayload(ctx, issue);

        // Отправляем webhook если есть изменения
        if (payload.changes && payload.changes.length > 0) {
            sendWebhook(payload);
        }
    },
    requirements: {
        Assignee: {type: entities.User.fieldType},
        State: {type: entities.EnumField.fieldType},
        Priority: {type: entities.EnumField.fieldType}
    }
});
