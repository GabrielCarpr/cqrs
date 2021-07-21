INSERT INTO users (id, name, email, hash, active, version) VALUES ('e67347d6-9a19-4bf0-83ed-fd62d2a53906', 'Gabriel Carpreau', 'me@gabrielcarpreau.com', '$2a$10$iRrRnyVD0EMLPWfsjieAAuFG1yGm31WeXQB1NVX2AJ13txwdJKUlG', true, 1);

INSERT INTO scopes (name) VALUES ('self:read'), ('self:write'), ('users:read'), ('users:write'), ('roles:read'), ('roles:write');

INSERT INTO roles (ID, label, version) VALUES ('admin', 'Admin', 1), ('user', 'User', 1);

INSERT INTO user_roles (user_id, role_id) VALUES ('e67347d6-9a19-4bf0-83ed-fd62d2a53906', 'user'), ('e67347d6-9a19-4bf0-83ed-fd62d2a53906', 'admin');

INSERT INTO role_scopes (role_id, scope_id) VALUES ('admin', 'self:read'), ('admin', 'users:read'), ('admin', 'users:write'), ('admin', 'self:write'), ('admin', 'roles:read'), ('admin', 'roles:write');
INSERT INTO role_scopes (role_id, scope_id) VALUES ('user', 'self:read'), ('user', 'self:write');

INSERT INTO payment_statuses (name)
    VALUES ('begin'), ('waiting'), ('requires_action'), ('succeeded'), ('failed');
