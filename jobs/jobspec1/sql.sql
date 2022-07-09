CREATE OR REPLACE FUNCTION fn_import_missing_vaccinatie_events (part integer, label  varchar, batchsize  integer)
RETURNS void
AS $$
  DECLARE
    cnt integer := 1;
    filter varchar := '';
  BEGIN

    EXECUTE format('CREATE TEMPORARY TABLE batch as SELECT ora_id FROM dba_check.%s LIMIT %s', quote_ident('missing_ora_ids_'||part::text), batchsize::text);

    SELECT INTO cnt count(*) FROM batch;
    WHILE cnt > 0 LOOP
      SELECT INTO filter string_agg(ora_id::text, ',') from batch;

      EXECUTE format('
        INSERT INTO public.%s (part_key, id, bsn_external, bsn_internal, payload, iv, version_cims, version_vcbe, created_at)
        SELECT mod(ora.id, 10), ora.id, ora.bsnextern, ora.bsnintern, ora.payload, ora.initialisatie_vector, ora.version, %s, clock_timestamp()
        FROM ora_vcbe.vaccinatie_event ora
        WHERE ora.id IN (%s)
        ', quote_ident('vaccinatie_event_'||part::text), quote_nullable(label), filter);

      EXECUTE format('DELETE FROM dba_check.%s WHERE ora_id IN (SELECT ora_id FROM batch)', quote_ident('missing_ora_ids_'||part::text));

      TRUNCATE batch;

      EXECUTE format('INSERT INTO batch SELECT DISTINCT ora_id FROM dba_check.%s LIMIT %s', quote_ident('missing_ora_ids_'||part::text), batchsize);

      SELECT INTO cnt count(*)
      FROM batch;

    END LOOP;
    DROP TABLE batch;
  END;
$$ LANGUAGE plpgsql;
